// @ts-check


let currentPath = window.location.pathname + window.location.search;

window.addEventListener("popstate", () => {
    const activePath = window.location.pathname + window.location.search;

    if (activePath !== currentPath) {
        currentPath = activePath;
        window.location.reload();
    }
});

document.addEventListener('click', (e) => {
    if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
    const anchor = (e.target instanceof HTMLAnchorElement) ? e.target.closest('a[data-dialog]') : null;
    if (!(anchor instanceof HTMLAnchorElement)) {
        return;
    }
    const dialog = document.getElementById(anchor.dataset.dialog ?? '')
    if (dialog instanceof HTMLDialogElement) {
        e.preventDefault();
        dialog.showModal();
    }
});

/**
 * 
 * @param {string} s 
 * @param {...(string)} args - Values to replace %s
 * @returns {string}
 */
function replacePlaceholders(s, ...args) {
    for (const arg of args) {
        s = s.replace(/%[sd]/, arg);
    }
    return s
}

//shared
/**
 * @typedef {HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement} PossiblyInvalidFormControl
 * 
 * An error of a formcontrol (which should only be made visible on formsubmit)
 * @typedef {Object} FormControlError
 * @property {string} id - id of the invalid formcontrol
 * @property {string} reason - (to be shown) errormessage
 * @property {keyof typeof validity|undefined} code - type of error
 * @property {Record<string, any>} errors
 */

/**
 * only yet needed subset of possible codes
 */
const validity = {
    required: 'required',
    minWords: 'minWords',
    maxCodePoints: 'maxCodePoints',
    maxBytes: 'maxBytes',
    invalid: 'invalid',
}

/**
 * @param {PossiblyInvalidFormControl} control
 * @param {HTMLElement|null} errorSummary
 */
function setValid(control, errorSummary) {
    const errorEl = getErrorMessageElement(control);

    if (errorEl) {
        const lastNode = errorEl.lastChild;

        if (lastNode && lastNode.nodeType === Node.TEXT_NODE) {
            lastNode.replaceWith(getErrorMessageDefault(control));
        } else {
            errorEl.append(getErrorMessageDefault(control));
        }
    }

    let idRemove = control.id
    if (control instanceof HTMLInputElement && control.type === 'radio') {
        control.closest('fieldset')?.toggleAttribute('data-invalid', false)
        idRemove = control.closest('fieldset')?.querySelector('input[type=radio]')?.id ?? idRemove
    }

    control.removeAttribute('data-invalid')
    control.ariaInvalid = 'false';
    if (errorSummary instanceof HTMLElement) {
        removeSummaryEntry(idRemove, errorSummary);
    }
}

/**
 * Gets the corresponding Element showing the errormessage if the control is aria-invalid
 * @param {PossiblyInvalidFormControl} control
 * @returns {HTMLElement|null|undefined}
 */
function getErrorMessageElement(control) {
    const describedBy = control.getAttribute('aria-describedby') || control.closest('fieldset')?.getAttribute('aria-describedby');
    if (!describedBy) return null;

    const ids = '#' + describedBy.trim().split(/\s+/).join(',#');

    return control.closest('.form-group')?.querySelector(`:is(${ids}) > .error-message`)
        || control.closest('fieldset')?.querySelector(`:is(${ids}) > .error-message`)
}


/**
 * Gets the default errormessage (most often for required) of the formcontrol
 * @param {PossiblyInvalidFormControl} control 
 * @returns {string}
 */
function getErrorMessageDefault(control) {
    let errorMessage = '';
    let errorElement = null;
    if ((errorElement = getErrorMessageElement(control)) instanceof HTMLElement) {
        errorMessage = errorElement.dataset.defaultMessage ?? '';
        if (errorMessage === '') {
            const lastNode = errorElement.lastChild;
            const text = (lastNode && lastNode.nodeType === Node.TEXT_NODE) ? (lastNode.nodeValue || '') : '';
            errorMessage = text.trim();
            errorElement.dataset.defaultMessage = errorMessage;
        }
    }
    return errorMessage;
}

/**
 * @param {string} id
 * @param {HTMLElement} errorSummary
 */
function removeSummaryEntry(id, errorSummary) {
    const listItem = errorSummary?.querySelector(`li:has(a[href="#${id}"])`);
    if (errorSummary instanceof HTMLElement && listItem instanceof HTMLLIElement) {
        listItem.remove();
        if (errorSummary.querySelector(`li`) === null) {
            errorSummary.hidden = true;
            clearTitleErrorPrefix();
        }
    }
}

/**
 * Gets the errormessage of the formcontrol
 * @param {PossiblyInvalidFormControl} control 
 * @returns {string}
 */
function getErrorMessage(control) {
    let errorMessage = '';
    let errorElement = null;
    if (control.validity.customError) {
        errorMessage = control.validationMessage;
    } else if ((errorElement = getErrorMessageElement(control)) instanceof HTMLElement) {
        const lastNode = errorElement.lastChild;
        if (lastNode && lastNode.nodeType === Node.TEXT_NODE && lastNode.nodeValue) {
            errorMessage = lastNode.nodeValue.trim();
        }
    }
    return errorMessage;
}

/**
 * If the formcontrol can be subject to constraint validation
 * @param {EventTarget | null} el 
 * @returns {el is PossiblyInvalidFormControl}
 */
function isValidateableElement(el) {
    if (el instanceof HTMLSelectElement || el instanceof HTMLTextAreaElement) {
        return true;
    }
    if (el instanceof HTMLInputElement) {
        return !['hidden', 'reset', 'button'].includes(el.type);
    }
    return false;
}



/**
 * 
 * @param {string} prefix 
 */
function setTitleErrorPrefix(prefix) {
    const titleElement = document.querySelector('title');
    if (!titleElement || 'error' in titleElement.dataset) {
        return;
    }
    if (prefix !== '') {
        document.title = `${prefix}: ${document.title}`;
    } else {
        document.title = `Error: ${document.title}`;
    }
    titleElement.dataset.error = prefix;
}


function clearTitleErrorPrefix() {
    const titleElement = document.querySelector('title');
    const prefix = titleElement?.dataset.error;
    if (!titleElement || !prefix) {
        return;
    }

    if (document.title.startsWith(`${prefix}: `)) {
        document.title = document.title.slice(prefix.length + 2);
    }

    delete titleElement.dataset.error;
}

//-- enhanced-form
{

    //server problem+json response shapes

    /**
     * @typedef {object} InvalidParam
     * @property {string} name
     * @property {string} code
     * @property {object} errors
     * @property {string} reason
     */

    /**
     * @typedef {object} ValidationProblem
     * @property {"/docs/problems/validation"} type
     * @property {string} title
     * @property {InvalidParam[]} errors
     */

    /**
     * @typedef {object} GenericProblem
     * @property {string} detail
     */


    /** @param {unknown} v @returns {v is InvalidParam} */
    function isProblemError(v) {
        return !!v && typeof v === 'object'
            && 'name' in v && typeof v.name === 'string'
            && 'code' in v && typeof v.code === 'string'
            && 'reason' in v && typeof v.reason === 'string'
            && (!('params' in v) || 'params' in v && typeof v.params === 'object');
    }

    /** @param {unknown} v @returns {v is ValidationProblem} */
    function isValidationProblem(v) {
        return !!v && typeof v === 'object'
            && 'errors' in v
            && Array.isArray(v['errors'])
            && v['errors'].every(isProblemError);
    }

    /** @param {unknown} v @returns {v is GenericProblem} */
    function isGenericProblem(v) {
        return !!v && typeof v === 'object'
            && 'detail' in v
            && typeof v.detail === 'string'
    }

    //-- client-side type and subset

    /**
     *
     * @typedef {HTMLInputElement | HTMLTextAreaElement} PossiblyBlank
     */


    const mediatypeHTML = 'text/html;charset=utf-8';
    const mediaTypeProblemJSON = 'application/problem+json';

    /**
     * 
     * @param {string} code 
     * @returns {code is keyof validity}
     */
    function isValidityCode(code) {
        return code in validity;
    }

    /**
     * 
     * @param {ValidationProblem} problem 
     * @param {HTMLFormElement} form
     * @returns {FormControlError[]}
     */
    function problemtoFormControlErrors(problem, form) {
        /** @type {FormControlError[]} */
        const errors = [];

        for (const error of problem.errors) {
            const id = getIdFromName(form, error.name);
            const code = error.code
            if (id !== null && isValidityCode(code)) {
                errors.push({
                    id: id,
                    code: code,
                    errors: error.errors,
                    reason: error.reason
                })
            }
        }


        return errors;
    }

    /**
     * 
     * @param {HTMLFormElement} form 
     * @param {string} name 
     * @returns {string|null}
     */
    function getIdFromName(form, name) {
        const control = form.elements.namedItem(name);
        if (control === null) {
            return null;
        }
        if (control instanceof RadioNodeList) {
            return control[0].id;
        }

        return control.id;
    }




    class EnhancedForm extends HTMLElement {

        /**
         * @type {HTMLFormElement|undefined}
         */
        #form

        /**
         * banner shown on direct navigation
         * @type {HTMLElement|undefined}
         */
        #succesBanner;

        /**
         * banner shown on non-field-specific error
         * @type {HTMLElement|undefined}
         */
        #errorBanner;

        /**
        * current shown banner
        * @type {HTMLElement|undefined}
        */
        #currBanner;

        /**
         * @type {HTMLLIElement|undefined}
         */
        #preconditionFailedListItem

        /**
         * @type {HTMLButtonElement|undefined}
         */
        #preconditionFailedSubmitButton

        /**
         * @type {string|undefined}
         */
        #errorTitlePrefix

        /**
         * error-summary shown for field-specific errors
         * @type {HTMLElement|undefined}
         */
        #errorSummary

        #init = false;

        #firstSubmit = true;
        /**
         * @type {AbortController|null}
         */
        #abortController = null;


        connectedCallback() {
            if (this.#init === true) {
                return;
            }
            const form = this.querySelector('form');
            const errorSummary = document.getElementById('error-summary');
            const hasPrecondition = form?.elements?.namedItem('If-Match') != null;
            const preconditionFailedListItem = /** @type {HTMLTemplateElement | null} */ (this.querySelector('template[data-precondition-list]'))?.content.firstElementChild?.cloneNode(true)
            const preconditionFailedSubmitButton = /** @type {HTMLTemplateElement | null} */ (this.querySelector('template[data-precondition-submit]'))?.content.firstElementChild?.cloneNode(true)
            const successBanner = /** @type {HTMLTemplateElement | null} */ (this.querySelector('template[data-success]'))?.content.firstElementChild?.cloneNode(true)
            const errorBanner = /** @type {HTMLTemplateElement | null} */ (this.querySelector('template[data-error]'))?.content.firstElementChild?.cloneNode(true);
            const errorTitlePrefix = this.getAttribute('error-title-prefix');

            if (!(form instanceof HTMLFormElement)
                || !(errorSummary instanceof HTMLElement)
                || !(successBanner instanceof HTMLElement)
                || !(errorBanner instanceof HTMLElement)
                || !errorTitlePrefix) {
                return;
            }
            if (hasPrecondition) {
                if (preconditionFailedListItem instanceof HTMLLIElement && preconditionFailedSubmitButton instanceof HTMLButtonElement) {
                    this.#preconditionFailedListItem = preconditionFailedListItem;
                    this.#preconditionFailedSubmitButton = preconditionFailedSubmitButton;
                } else {
                    return;
                }
            }
            this.#init = true;
            this.#firstSubmit = true;
            this.#form = form;
            this.#errorSummary = errorSummary;
            this.#succesBanner = successBanner;
            this.#errorBanner = errorBanner;
            this.#errorTitlePrefix = errorTitlePrefix;

            this.#abortController = new AbortController();
            const { signal } = this.#abortController;

            this.#form.addEventListener('submit', this.#enhanceSubmit, { signal })
        }

        disconnectedCallback() {
            this.#abortController?.abort();
        }

        /**
         * 
         * @param {SubmitEvent} e 
         * @returns 
         */
        #enhanceSubmit = async (e) => {

            if (!this.#form || !this.#errorSummary) {
                return;
            }

            const errors = collectErrors(this.#form)
            if (errors.length > 0) {
                e.preventDefault();
                if (this.#firstSubmit) {
                    this.#applyEarlyRewards();
                    this.#firstSubmit = false;
                }
                applyErrors(errors, this.#errorSummary, this.#errorTitlePrefix ?? '');
                return;
            }

            e.preventDefault();

            const formdata = new FormData(this.#form, e.submitter);
            //source is either 'keep' or decides between different fields of which js disabled the ones of unchecked radios anyway
            const source = formdata.get('source')
            if (typeof source === 'string' && formdata.has(source)) {
                formdata.delete('source');
            }

            const preference = this.getAttribute('prefer-return') === 'representation' ? 'representation' : 'minimal';
            const ifMatch = formdata.get('If-Match');

            /** @type {Record<string, string>} */
            const headers = {
                'Accept': `${mediaTypeProblemJSON},${mediatypeHTML}`,
                'Prefer': `return=${preference}`,
            };

            if (typeof ifMatch === 'string' && ifMatch.trim() !== '') {
                headers['If-Match'] = ifMatch.trim();
            }

            let response = null;
            try {
                response = await fetch(this.#form.action, {
                    body: formdata,
                    headers: headers,
                    method: this.#form.method,
                })
            } catch (error) {
                this.#handleSubmitError(null, this.#errorTitlePrefix ?? '');
                return
            }

            if (!response.ok) {
                this.#handleSubmitError(response, this.#errorTitlePrefix ?? '');
                return;
            }

            const url = response.headers.get('Content-Location')
                ?? response.headers.get('Location')
                ?? response.url;


            const isHTML = getBaseMediaType(response.headers.get('Content-Type') ?? '') === 'text/html';

            if (preference !== 'representation' || !isHTML) {
                await response.body?.cancel();
                window.location.assign(url);
                return;
            }

            await this.#navigate(response, url);
        }

        /**
         */
        #applyEarlyRewards() {
            /**
             * 
             * @param {Event} e 
             */
            const onInput = (e) => {
                if (this.#errorSummary instanceof HTMLElement) {
                    reward(e, this.#errorSummary);
                }
            };
            const signal = this.#abortController?.signal;
            if (this.#form) {
                this.#form.addEventListener('input', onInput, { signal });
                this.#form.addEventListener('change', onInput, { signal });
            }
        }

        /**
         * @param {Response|null} response
         * @param {string} titleErrorPrefix
         * @returns
         */
        async #handleSubmitError(response, titleErrorPrefix) {
            let problem = null
            if (response && getBaseMediaType(response.headers.get('Content-Type') ?? '') === mediaTypeProblemJSON) {
                try { problem = await response.json(); } catch { /* fall through */ }
                if (isValidationProblem(problem) && this.#form && this.#errorSummary) {
                    const errors = problemtoFormControlErrors(problem, this.#form);
                    if (errors.length > 0) {
                        applyErrors(errors, this.#errorSummary, titleErrorPrefix);
                        return;
                    }
                }
                if (response.status === 412
                    && this.#form && this.#errorSummary
                    && this.#preconditionFailedListItem && this.#preconditionFailedSubmitButton) {
                    await this.#applyPreconditionFailed(response.url);
                    return;
                }
            }

            if (problem !== null && typeof problem === 'object' && !Array.isArray(problem)
                && typeof problem.detail === 'string' && problem.detail.length > 0) {
                this.#showErrorBanner(problem.detail, this.#errorTitlePrefix ?? '');
                return;
            }

            this.#showErrorBanner('', this.#errorTitlePrefix ?? '');
        }

        /**
         * @param {string} url
         */
        async #applyPreconditionFailed(url) {
            const formSubmitButton = this.#form?.querySelector('button:is([type="submit"], :not([type])):not([form],[formmethod])');

            if (!(formSubmitButton instanceof HTMLButtonElement
                && this.#form
                && this.#errorSummary
                && this.#preconditionFailedListItem
                && this.#preconditionFailedSubmitButton)) {
                return
            }

            let newEtag = '';
            try {
                const response = await fetch(url, {
                    method: 'HEAD',
                    headers: { 'Accept': 'application/json' }
                })
                if (!response.ok) {
                    this.#form.submit();
                    return;
                }
                newEtag = response.headers.get('ETag') ?? ''
            } catch (error) {
                this.#form.submit();
                return;
            }
            if (newEtag === '') {
                this.#form.submit();
                return;
            }
            formSubmitButton.replaceWith(this.#preconditionFailedSubmitButton);
            const currIfMatch = this.#form.elements.namedItem('If-Match');
            if (currIfMatch instanceof HTMLInputElement) {
                currIfMatch.value = newEtag;
            } else {
                const newIfMatch = document.createElement('input');
                newIfMatch.type = 'hidden';
                newIfMatch.name = 'If-Match'
                newIfMatch.value = newEtag;
                this.#form.prepend(newIfMatch);
            }
            applyLi(this.#preconditionFailedListItem, this.#errorSummary);
        }


        /**
         * 
         * @param {Response} response 
         * @param {string} location 
         */
        async #navigate(response, location) {
            let html = await response.text();
            const dom = (new DOMParser()).parseFromString(html, "text/html");
            if (this.#succesBanner) {
                dom.querySelector('main')?.prepend(this.#succesBanner)
            }
            updateHead(dom.head);
            document.body.replaceChildren(...dom.body.children);
            history.pushState(null, '', location);
            currentPath = location;
            scrollAndFocus();
        }

        /**
         * 
         * @param {string} detail
         * @param {string} titleErrorPrefix
         */
        #showErrorBanner(detail, titleErrorPrefix) {
            if (!this.#errorBanner) {
                return
            }
            const Banner = this.#errorBanner.cloneNode(true);
            if (!(Banner instanceof HTMLElement)) {
                return;
            }
            if (detail) {
                const elementReplace = Banner.querySelector('[data-replace]')
                if (elementReplace) {
                    elementReplace.textContent = detail;
                }
            }
            if (this.#currBanner) {
                this.#currBanner.remove();
            }
            this.#currBanner = Banner;
            document.querySelector('main')?.prepend(Banner);
            setTitleErrorPrefix(titleErrorPrefix);
            Banner.focus();
        }

    }
    customElements.define('enhanced-form', EnhancedForm);


    /**
     * 
     * @param {HTMLLIElement} li 
     * @param {HTMLElement} errorSummary 
     */
    function applyLi(li, errorSummary, errorTitlePrefix = 'Error') {
        if (errorSummary instanceof HTMLElement) {
            errorSummary.querySelector('ul')?.replaceChildren(li);
            errorSummary.hidden = false;
            setTitleErrorPrefix(errorTitlePrefix);
            errorSummary.focus();
        }
    }

    /**
     * sets formcontrols with errors aria-invalid, creates and inserts or updates the error-summary and focuses it
     * @param {FormControlError[]} errors
     * @param {HTMLElement|null} errorSummary
     */
    function applyErrors(errors, errorSummary, errorTitlePrefix = 'Error') {
        /**
         * @type {Node[]}
         */
        let newErrorItems = [];
        for (const error of errors) {
            const control = document.getElementById(error.id);
            if (!isValidateableElement(control)) {
                continue;
            }
            let errAnchor = document.createElement('a');
            errAnchor.href = '#' + error.id;
            errAnchor.textContent = error.reason;
            let li = document.createElement('li');
            li.append(errAnchor);
            newErrorItems.push(li);
            applyError(error, control);
        }

        if (errorSummary instanceof HTMLElement) {
            errorSummary.querySelector('ul')?.replaceChildren(...newErrorItems);
            setTitleErrorPrefix(errorTitlePrefix);
            errorSummary.hidden = false;
            errorSummary.focus();
        }
    }

    /**
     * 
     * @param {FormControlError} error 
     * @param {PossiblyInvalidFormControl} control
     */
    function applyError(error, control) {

        const errorEl = getErrorMessageElement(control);

        if (errorEl) {
            const lastNode = errorEl.lastChild;

            if (lastNode && lastNode.nodeType === Node.TEXT_NODE) {
                lastNode.replaceWith(error.reason);
            }
        }

        if (error.code !== undefined) {
            control.dataset.invalid = error.code;
        }

        if (control instanceof HTMLInputElement && control.type === 'radio') {
            control.closest('fieldset')?.toggleAttribute('data-invalid', true)
        } else {
            control.ariaInvalid = 'true';
        }
    }


    //-- notification banner


    /**
    * 
    * @param {HTMLFormElement} form 
    */
    function collectErrors(form) {
        // arm required on active conditional controls
        form.querySelectorAll('conditional-reveal > fieldset :is(input, textarea, select):not(:disabled)[aria-required="true"]').forEach((control) => {
            if (control instanceof HTMLInputElement || control instanceof HTMLTextAreaElement || control instanceof HTMLSelectElement) {
                control.required = true;
            }
        });
        /** @type {FormControlError[]} */
        const errors = [];

        const seenRadioNames = new Set();

        const radioSeen = ( /** @type {PossiblyInvalidFormControl} */ control) => {
            if (control instanceof HTMLInputElement && control.type === 'radio' && seenRadioNames.has(control.name)) {
                return true;
            }
            seenRadioNames.add(control.name);
            return false;
        }

        for (const control of form.elements) {
            if (!isValidateableElement(control)
                || !control.willValidate
                || radioSeen(control)) {
                continue;
            }
            let error = null;
            error = validateWordCount(control)
                || validateCodePoints(control)
                || validateFile(control)
                || validateNotBlank(control);

            if (error !== null) {
                errors.push(error);
            } else if (!control.validity.valid) {
                errors.push({
                    id: control.id,
                    code: undefined,
                    reason: getErrorMessageDefault(control),
                    errors: {}
                })
            }
        }

        return errors
    }

    //-- validation
    //-- wordcounts
    const selectorMinWords = 'textarea[data-minwords][data-minwords-error]:not(:is(:disabled, :read-only))';

    /**
     * 
     * @param {Element} control 
     * @returns {control is HTMLTextAreaElement}
     */
    function hasMinWords(control) {
        return control.matches(selectorMinWords);
    }

    /**
     * 
     * @param {Element} control
     * @returns {FormControlError|null}
     */
    function validateWordCount(control) {
        if (!hasMinWords(control)) {
            return null;
        }
        const min = parseInt(control.dataset.minwords || '', 10) || 0;
        if (min <= 0) return null;

        const count = (control.value.match(/\S+/g) || []).length;

        if (count < min && (count > 0 || control.required)) {
            let message = control.dataset.minwordsError ?? '';
            message = replacePlaceholders(message, min.toLocaleString());
            return {
                id: control.id,
                code: 'minWords',
                reason: message,
                errors: { limit: min }
            };
        }
        return null;
    }

    //-- codepoints
    /**
     * 
     * @param {Element} control 
     * @returns {control is HTMLInputElement}
     */
    function hasMaxCodePoints(control) {
        return control.matches('input[type="text"][data-max-code-points][data-max-code-points-error]:not(:is(:disabled, :read-only))');
    }

    /**
     * 
     * @param {Element} control 
     * @returns {FormControlError|null}
     */
    function validateCodePoints(control) {
        if (!hasMaxCodePoints(control)) {
            return null;
        }
        const max = parseInt(control.dataset.maxCodePoints || '', 10) || 0;
        if (max <= 0) return null;

        if ([...control.value].length > max && (control.value.length > 0 || control.required)) {
            let message = control.dataset.maxCodePointsError ?? '';
            message = replacePlaceholders(message, max.toLocaleString());
            return {
                id: control.id,
                code: 'maxCodePoints',
                reason: message,
                errors: { limit: max }
            };
        }
        return null;
    }

    //-- file validity
    /**
     * 
     * @param {Element} control 
     * @returns {control is HTMLInputElement}
     */
    function isValidateableFile(control) {
        return control.matches('input[type="file"]:where([data-max-bytes][data-max-bytes-error], required, [accept*=".txt"][accept*="text/plain"]):not(:disabled)');
    }

    /**
     * 
     * @param {Element} control 
     * @returns {FormControlError|null}
     */
    function validateFile(control) {
        if (!isValidateableFile(control)) {
            return null;
        }

        if (control.validity.valueMissing) {
            return {
                id: control.id,
                code: 'required',
                reason: getErrorMessageDefault(control),
                errors: {}
            }
        }

        const file = control.files?.[0];
        if (!file) {
            return null;
        }

        let currentFileKey = '';

        if (file instanceof File && (currentFileKey = `${file.name}-${file.size}-${file.lastModified}`) !== control.dataset.key) {
            if (control.dataset.maxBytes && control.dataset.maxBytesError
                && file.size > Number(control.dataset.maxBytes)) {
                return {
                    id: control.id,
                    code: 'maxBytes',
                    reason: control.dataset.maxBytesError,
                    errors: { limit: control.dataset.maxBytes }
                }
            } else if (control.matches('[accept*=".txt"][accept*="text/plain"]') && control.dataset.notTxtError && (file.type !== "text/plain" || !file.name.endsWith('.txt'))) {
                return {
                    id: control.id,
                    code: 'invalid',
                    reason: control.dataset.notTxtError,
                    errors: {}
                }
            }
            control.dataset.key = currentFileKey;
        }
        return null;
    }

    //-- blank

    const selectorBlank = ':is(input[type="text"], textarea)[data-rules$="notBlank"]:valid:not(:is(:disabled, :read-only))';

    /**
     * 
     * @param {Element} control 
     * @returns {control is HTMLInputElement|HTMLTextAreaElement}
     */
    function canBeBlank(control) {
        return control.matches(selectorBlank);
    }

    /**
     * returns true if not blank, false (+sets error message) if is.
     * can only be blank if not suffering from any other (custom) validity states
     * @param {HTMLElement} control
     * @returns {FormControlError|null}
     */
    function validateNotBlank(control) {
        if (!canBeBlank(control)) {
            return null;
        }
        let defaultMessage = getErrorMessageDefault(control);
        if (!control.validity.customError || getErrorMessage(control) === defaultMessage) {
            if (control.value.trim() === '') {
                return {
                    id: control.id,
                    code: 'required',
                    reason: defaultMessage,
                    errors: { rule: 'notBlank' }
                };
            }
        }
        return null;
    }

    /**
     * @param {Event} e
     * @param {HTMLElement | null} errorSummary
     */
    function reward(e, errorSummary) {

        const control = e.target
        let validityState = undefined;
        if (!isValidateableElement(control)
            || control.ariaInvalid === 'false'
        ) {
            return;
        }

        let valid = false;

        validityState = control.dataset.invalid ?? ''

        switch (validityState) {
            case 'minWords':
                valid = validateWordCount(control) === null
                break;
            case 'maxCodePoints':
                valid = validateCodePoints(control) === null
                break;
            case 'invalid':
                valid = validateFile(control) === null
                break;
            default:
                valid = control.validity.valid && validateNotBlank(control) === null
                break;
        }


        if (valid) {
            setValid(control, errorSummary)
        }
    }
}


//-- radios with conditional reveal
class ConditionalReveal extends HTMLElement {
    /** @type {boolean} */
    #init = false;

    /**
     * @type {AbortController|null}
     */
    #abortController = null;


    connectedCallback() {
        if (this.#init) {
            return;
        }
        this.#init = true;

        this.#abortController = new AbortController();
        const { signal } = this.#abortController;

        //arm required only after the user interacted (to not prematurely sho errors since native formsubmit sets even disabled formcontrols user touched on submit, turning required ones even if disabled invalid once enabled)
        this.addEventListener('blur', (event) => {
            const control = event.target;
            if ((control instanceof HTMLInputElement || control instanceof HTMLTextAreaElement || control instanceof HTMLSelectElement)
                && !control.required
                && !control.disabled
                && control.getAttribute('aria-required') === 'true'
            ) {
                control.required = true;
            }
        }, { signal: signal, capture: true });

        this.addEventListener('change', (event) => {
            if (!(event.target instanceof HTMLInputElement)
                || event.target.type !== 'radio') {
                return;
            }
            this.sync()
        }, { signal });
        this.sync()
    }

    sync() {
        /** @type {NodeListOf<PossiblyInvalidFormControl>}*/
        const controlsDisable = this.querySelectorAll('.form-group:has(input[type="radio"]:not(:checked)) + * :is(input,textarea,select):not(:disabled)')
        const errorSummary = document.getElementById('error-summary')
        controlsDisable.forEach((control) => {
            control.disabled = true;
            if (errorSummary instanceof HTMLElement) {
                setValid(control, errorSummary);
            }
        })

        /** @type {NodeListOf<PossiblyInvalidFormControl>}*/
        const controlsEnable = this.querySelectorAll('.form-group:has(input[type="radio"]:checked) + * :is(input,textarea,select):disabled')
        controlsEnable.forEach((control) => {
            control.disabled = false;
        })
    }
}
customElements.define('conditional-reveal', ConditionalReveal);


//-- word and character count
{
    /**
     * 
     * @param {string} text
     * @returns {number}
     */
    function countWords(text) {
        const matches = text.match(/\S+/g);
        return matches ? matches.length : 0;
    }

    /**
     * counts words and, if limit is given, adds advisory message to count when above it
     */
    class WordCountAdvisory extends HTMLElement {
        /** @type {boolean} */
        #init = false;

        /** @type {HTMLTextAreaElement|null} */
        #textarea = null;

        /** @type {HTMLElement|null} */
        #visual = null;

        /** @type {HTMLElement|null} */
        #announce = null;

        /** @type {string} */
        #one = '';

        /** @type {string} */
        #other = '';

        /** @type {string|undefined} */
        #over;

        /** @type {number|undefined} */
        #limit;

        /**
         * @type {number | undefined}
         */
        #anounceDebounce;

        /**
         * @type {AbortController|null}
         */
        #abortController = null;


        connectedCallback() {
            if (this.#init) {
                return;
            }

            const textarea = this.querySelector('textarea');
            const textAreaID = textarea?.id ?? '';
            const info = this.querySelector(`[id="${textAreaID}-info"]`);
            const visual = this.querySelector('[aria-hidden="true"]');
            const announce = this.querySelector('[aria-live="polite"].visually-hidden')
            const one = this.getAttribute('eins');
            const other = this.getAttribute('other');

            if (!textarea || !(info instanceof HTMLElement) || !(visual instanceof HTMLElement) || !(announce instanceof HTMLElement) || !one || !other) {
                return;
            }

            this.#init = true;
            this.#textarea = textarea;
            this.#visual = visual;
            this.#announce = announce;
            this.#one = one;
            this.#other = other;
            this.#abortController = new AbortController();
            const { signal } = this.#abortController;

            const limit = this.getAttribute('limit');
            if (limit !== null) {
                const parsed = Number.parseInt(limit, 10);
                if (!Number.isInteger(parsed) || parsed <= 0) {
                    return;
                }
                this.#limit = parsed;
            }
            const over = this.getAttribute('over');
            if (over) {
                this.#over = over;
            }

            this.#update();
            info.hidden = true;
            this.#textarea.addEventListener('input', this.#update, { signal })
        }

        disconnectedCallback() {
            clearTimeout(this.#anounceDebounce);
            this.#abortController?.abort();
            this.#init = false;
        }

        #update = () => {
            if (!this.#textarea || !this.#visual) {
                return;
            }

            const count = countWords(this.#textarea.value);

            let message = "";

            message = this.#getMessage(count);

            // Visual update (instant)
            this.#visual.textContent = message;

            // Screen reader update (debounced, empty if below threshold)
            clearTimeout(this.#anounceDebounce);
            this.#anounceDebounce = setTimeout(() => {
                if (this.#announce) {
                    this.#announce.textContent = message;
                }
            }, 1000);
        }

        /**
         * 
         * @param {number} count
         * @returns {string}
         */
        #getMessage(count) {

            let message = "";
            if (count === 1) {
                message = this.#one;
            } else {
                message = replacePlaceholders(this.#other, String(count));
            }

            if (this.#limit && this.#limit - count < 0) {
                message = `${message} - ${this.#over}`
            }

            return message;
        }

    }
    customElements.define('word-count-advisory', WordCountAdvisory);
}

//-- enhanced navigation

/**
 * 
 * @param {string} header 
 * @returns 
 */
function getBaseMediaType(header) {
    if (!header) return '';
    return header.split(';')[0].trim().toLowerCase();
}


/**
 * Updates the html head based on a new head.
 * Currently only updates metadata (no styles or scripts)
 * @param {HTMLHeadElement} newHead
 */
function updateHead(newHead) {
    if (!(newHead instanceof HTMLHeadElement)) {
        return;
    }
    // update title
    let newTitleElement = null;
    if ((newTitleElement = newHead.querySelector('title')) instanceof HTMLTitleElement) {
        document.title = newTitleElement.textContent;
    }
    // update meta-elements
    /**
     * Helper to get the "key" of a meta element
     * @param {HTMLMetaElement} el 
     * @returns {string|null}
     */
    const getKey = (el) =>
        el.getAttribute("name") ||
        el.getAttribute("property") ||
        el.getAttribute("http-equiv");

    const oldMeta = new Map();
    for (const meta of document.head.querySelectorAll("meta")) {
        const key = getKey(meta);
        if (key) {
            oldMeta.set(key, meta);
        }
    }

    for (const newMeta of newHead.querySelectorAll("meta")) {
        const key = getKey(newMeta);
        if (!key) continue; // Ignore charset, viewport, ...

        if (oldMeta.has(key)) {
            const existing = oldMeta.get(key);
            const newContent = newMeta.getAttribute("content");
            if (existing.getAttribute("content") !== newContent) {
                existing.setAttribute("content", newContent);
            }
            oldMeta.delete(key);
        } else {
            document.head.appendChild(newMeta.cloneNode(true));
        }
    }

    for (const [_, deadMeta] of oldMeta) {
        deadMeta.remove();
    }

    let linksNotSacred = 'link:not(:is([rel="stylesheet"],[rel="icon"],[rel="apple-touch-icon"]))';

    //remove outdated link(-rel)s
    document.querySelectorAll(linksNotSacred).forEach(element => {
        element.remove();
    });

    //append new link(-rel)s
    document.head.append(...newHead.querySelectorAll(linksNotSacred))

    return;
}

//(by default focus srolls)
function scrollAndFocus() {
    const focusTarget = document.getElementById('error-summary')
        || document.querySelector("[autofocus]")
        || document.querySelector("h1")
        || document.querySelector("main");

    if (focusTarget instanceof HTMLElement) {
        if (!focusTarget.hasAttribute("tabindex")) {
            focusTarget.setAttribute("tabindex", "-1");
        }
        focusTarget.focus();
    } else {
        document.body.focus();
    }
}


//-- custom element text status
{
    /**
     * @typedef {Object} StatusSchema
     * @property {string} status
     * @property {string} message
     * @property {Number|undefined} progress
     * @property {string|undefined} stage
     */

    /**
     * 
     * @param {any} thing
     * @returns {thing is StatusSchema}
     */
    function isStateSchema(thing) {
        return !!thing && typeof thing === 'object'
            && 'status' in thing && typeof thing.status === 'string'
            && 'message' in thing && typeof thing.message === 'string'
            && (!('progress' in thing) || typeof thing.progress === 'number')
            && (!('stage' in thing) || typeof thing.stage === 'string');
    }

    const statePending = 'pending';
    const stateProcessing = 'processing';


    class TextStatus extends HTMLElement {

        constructor() {
            super();
            console.log('text-status constructor');
        }

        /**
         * @type {HTMLAnchorElement|undefined}
         */
        #anchorRefresh


        connectedCallback() {
            console.log('text-status connected start');
            if (this.initiated || !this.isConnected) {
                return;
            }

            const messageElement = this.querySelector('[data-key="message"]');
            const anchorFallback = this.querySelector('a[rel="status"]');
            const anchorRefresh = this.querySelector('a[href=""]');
            const anchorSuggestions = this.querySelector('[rel="search"]')
            const liveRegion = this.querySelector('[data-live]');
            const progress = this.querySelector('progress');
            const progressElement = this.querySelector('[data-key="progress"]');
            const announce = this.querySelector('.visually-hidden');

            if (!(messageElement instanceof HTMLElement)
                || !(anchorFallback instanceof HTMLAnchorElement)
                || !(anchorRefresh instanceof HTMLAnchorElement)
                || !(anchorSuggestions instanceof HTMLAnchorElement)
                || !(liveRegion instanceof HTMLElement)
                || !((progress instanceof HTMLProgressElement) || progress === null)
                || !((progressElement instanceof HTMLElement) || progressElement === null)
                || !(announce instanceof HTMLElement)
            ) {
                return;
            }
            this.initiated = true
            this.#anchorRefresh = anchorRefresh
            console.log('text-status connected initiated');
            let lastMessage = messageElement.textContent;
            const mode = this.getAttribute('mode') ?? 'poll';
            const elements = new Map(
                Array.from(
                    /** @type {NodeListOf<HTMLElement>} */(
                        this.querySelectorAll('[data-key]:not(:is([data-key="message"], [data-key="progress"], progress))')
                    ), el => [el.dataset.key ?? '', el])
            );


            this.liveWrapper = liveRegion;
            this.announce = announce;

            let changedSinceAnounce = false;



            const anounceInterval = setInterval(() => {
                if (changedSinceAnounce) {
                    announce.textContent = `${progress?.textContent} ${messageElement.textContent}`;
                    changedSinceAnounce = false;
                }
            }, 2000);
            this.anounceInterval = anounceInterval

            const stageElement = elements.get('stage');

            if (mode === 'sse') {
                const source = new EventSource(anchorFallback.href);
                this.source = source;

                source.onopen = () => {
                    console.log("sse open")
                }

                source.onmessage = (event) => {
                    console.log("data:" + event.data)
                    messageElement.textContent = event.data;
                };

                source.addEventListener('status', (event) => {
                    console.log("event: status");
                    console.log("data:" + event.data);
                    this.setAttribute('status', event.data);
                    if (event.data == 'completed') {
                        anchorSuggestions.hidden = false;
                    }
                })

                for (const [key, element] of elements) {
                    source.addEventListener(key, (event) => {
                        console.log("event:" + key)
                        console.log("data:" + event.data)
                        element.textContent = event.data;
                        lastMessage = event.data;
                    });
                }
                source.addEventListener('progress', (event) => {
                    console.log("event:progress")
                    console.log("data:" + event.data)
                    if (progressElement) {
                        progressElement.textContent = `${event.data} %`
                    }
                    if (progress !== null) {
                        progress.value = Number(event.data);
                        progress.textContent = `${event.data} %`;
                    }
                    liveRegion.hidden = false;
                    anchorFallback.hidden = true;
                    changedSinceAnounce = true;
                });

                const endSSE = (/** @type {string} */ message) => this.endSSE(source, message, messageElement, liveRegion, anounceInterval)

                source.addEventListener('done', (event) => {
                    console.log("event:done")
                    console.log("data:" + event.data)
                    endSSE(event.data);
                });


                source.onerror = async () => {
                    try {
                        const stateSnapshot = await fetch(anchorFallback.href, {
                            headers: {
                                "Accept": "application/json"
                            }
                        });
                        if (!stateSnapshot.ok) {
                            endSSE(lastMessage);
                            return;
                        }
                        const snapshotJSON = await stateSnapshot.json();
                        if (isStateSchema(snapshotJSON)) {
                            this.setAttribute("status", snapshotJSON.status);
                            if (snapshotJSON.status !== stateProcessing && snapshotJSON.status !== statePending) {
                                anchorFallback.hidden = true;
                                if (snapshotJSON.status == 'completed') {
                                    anchorSuggestions.hidden = false;
                                }
                                endSSE(snapshotJSON.message)
                            } else {
                                messageElement.textContent = snapshotJSON.message;
                                if (stageElement && snapshotJSON.stage) {
                                    stageElement.textContent = snapshotJSON.stage;
                                }
                                if (snapshotJSON.progress && progress) {
                                    const progressStr = snapshotJSON.progress.toString();
                                    if (progressElement) {
                                        progressElement.textContent = progressStr;
                                    }
                                    progress.textContent = progressStr;
                                    progress.value = snapshotJSON.progress;
                                    changedSinceAnounce = true;
                                    liveRegion.hidden = false;
                                    anchorFallback.hidden = true;
                                }
                            }
                        } else {
                            source.close();
                            clearInterval(anounceInterval);
                        }
                    } catch (error) {
                        source.close();
                        clearInterval(anounceInterval);
                    }
                }
            } else {
                try {
                    const pollInterval = setInterval(async () => {
                        try {
                            const stateSnapshot = await fetch(anchorFallback.href ?? '', {
                                headers: {
                                    "Accept": "application/json"
                                }
                            });
                            if (!stateSnapshot.ok) {
                                changedSinceAnounce = true;
                                this.setStatic(lastMessage, messageElement, liveRegion);
                                clearInterval(pollInterval);
                                clearInterval(anounceInterval);
                                return;
                            }
                            const snapshotJSON = await stateSnapshot.json();
                            if (isStateSchema(snapshotJSON)) {
                                lastMessage = snapshotJSON.message;
                                if (snapshotJSON.status !== stateProcessing && snapshotJSON.status !== statePending) {
                                    changedSinceAnounce = true;
                                    this.setStatic(snapshotJSON.message, messageElement, liveRegion);
                                    clearInterval(anounceInterval);
                                    clearInterval(pollInterval);
                                    return
                                }

                                if (snapshotJSON.message !== messageElement.textContent) {
                                    messageElement.textContent = snapshotJSON.message;
                                }
                                if (snapshotJSON.stage && stageElement && snapshotJSON.stage !== stageElement.textContent) {
                                    stageElement.textContent = snapshotJSON.stage;
                                }
                                if (snapshotJSON.progress && progress) {
                                    const progressStr = snapshotJSON.progress.toString();
                                    if (progressElement) {
                                        progressElement.textContent = progressStr;
                                    }
                                    progress.textContent = progressStr;
                                    progress.value = snapshotJSON.progress;
                                    changedSinceAnounce = true;
                                    liveRegion.hidden = false;
                                    anchorFallback.hidden = true;
                                }
                            } else {
                                clearInterval(anounceInterval);
                                clearInterval(pollInterval);
                            }
                        } catch (error) {
                            clearInterval(anounceInterval);
                            clearInterval(pollInterval);
                        }
                    }, 2000);
                    this.pollInterval = pollInterval;
                } catch (error) {
                    clearInterval(anounceInterval);
                }
            }

        }

        /**
         * @param {string} message
         * @param {HTMLElement} stateElement
         * @param {HTMLElement} liveRegion
         */
        setStatic(message, stateElement, liveRegion) {
            stateElement.textContent = message;
            liveRegion.remove();
        }

        /**
         * @param {EventSource} source
         * @param {string} message
         * @param {HTMLElement} stateElement
         * @param {HTMLElement} liveRegion
         * @param {number | undefined} interval
         */
        endSSE(source, message, stateElement, liveRegion, interval) {
            source.close();
            this.source = null;
            this.setStatic(message, stateElement, liveRegion);
            clearInterval(interval);
            if (this.#anchorRefresh) {
                this.#anchorRefresh.hidden = false;
            }
            console.log('text-status sse ended');
        }



        disconnectedCallback() {
            this.source?.close();
            this.source = null;
            clearInterval(this.anounceInterval);
            clearInterval(this.pollInterval);
            this.initiated = false;
            this.liveWrapper = null;
            this.announce = null;
            console.log('text-status disconnected');
        }

    }
    customElements.define('text-status', TextStatus);
}

//--custom element suggestions-enhance
{
    class SuggestionsEnhance extends HTMLElement {
        /** @type {HTMLFormElement|null} */
        #form = null;
        /** @type {HTMLTextAreaElement|null} */
        #textarea = null;
        /** @type {HTMLFieldSetElement|null} */
        #fieldset = null;
        /** @type {HTMLElement|null} */
        #buttonBar = null;
        /** @type {AbortController|null} */
        #fetchAbort = null;
        /** @type {number|undefined} */
        #debounceTimer;

        connectedCallback() {
            const form = this.querySelector('form');
            const textarea = this.querySelector('textarea[name="text"]');
            const fieldset = this.querySelector('fieldset');
            const buttonBar = fieldset?.querySelector('.action-bar');
            const submitBar = this.querySelector('form > .action-bar');

            if (!(form instanceof HTMLFormElement)
                || !(textarea instanceof HTMLTextAreaElement)
                || !(fieldset instanceof HTMLFieldSetElement)
                || !(buttonBar instanceof HTMLElement)) {
                return;
            }

            this.#form = form;
            this.#textarea = textarea;
            this.#fieldset = fieldset;
            this.#buttonBar = buttonBar;

            if (submitBar instanceof HTMLElement) {
                submitBar.hidden = true;
            }

            form.addEventListener('submit', this.#onSubmit);
            textarea.addEventListener('input', this.#onInput);
        }

        disconnectedCallback() {
            clearTimeout(this.#debounceTimer);
            this.#fetchAbort?.abort();
        }

        /** @param {SubmitEvent} e */
        #onSubmit = (e) => {
            if (!this.#textarea || !this.#form) {
                return;
            }
            if (e.submitter instanceof HTMLButtonElement && e.submitter.hasAttribute('formaction')) {
                return;
            }
            e.preventDefault();

            const submitter = e.submitter;
            if (submitter instanceof HTMLButtonElement && submitter.name === 'next') {
                const nextWord = submitter.value;
                // Trim trailing partial word (if cursor/text ends in letters/numbers), then append nextWord + space
                const current = this.#textarea.value;
                const base = current.replace(/[\p{L}\p{N}]+$/u, '');
                this.#textarea.value = base + nextWord + ' ';
                this.#textarea.focus();
                this.#textarea.setSelectionRange(this.#textarea.value.length, this.#textarea.value.length);
            }

            clearTimeout(this.#debounceTimer);
            this.#fetchSuggestions();
        };

        #onInput = () => {
            clearTimeout(this.#debounceTimer);
            this.#debounceTimer = setTimeout(() => {
                this.#fetchSuggestions();
            }, 150);
        };

        async #fetchSuggestions() {
            if (!this.#form || !this.#textarea || !this.#fieldset || !this.#buttonBar) return;

            this.#fetchAbort?.abort();
            this.#fetchAbort = new AbortController();

            const params = new URLSearchParams();
            const formData = new FormData(this.#form);
            for (const [k, v] of formData.entries()) {
                if (typeof v === 'string' && v !== '' && !(k === 'textID' && v === '0') && !(k === 'limit' && v === '5')) {
                    params.set(k, v);
                }
            }

            const queryStr = params.toString();
            const url = `/suggestions${queryStr ? '?' + queryStr : ''}`;
            history.replaceState(null, '', url);
            currentPath = url;

            try {
                const resp = await fetch(url, {
                    headers: { 'Accept': 'application/linkset+json' },
                    signal: this.#fetchAbort.signal,
                });
                if (!resp.ok) return;

                const data = await resp.json();
                const itemLinks = data?.linkset?.[0]?.item;
                const buttons = [];

                if (Array.isArray(itemLinks)) {
                    for (const link of itemLinks) {
                        if (typeof link?.title === 'string') {
                            const btn = document.createElement('button');
                            btn.type = 'submit';
                            btn.name = 'next';
                            btn.value = link.title;
                            btn.textContent = link.title;
                            buttons.push(btn);
                        }
                    }
                }

                this.#buttonBar.replaceChildren(...buttons);
                this.#fieldset.hidden = buttons.length === 0;
            } catch (err) {
                if (err instanceof DOMException && err.name === 'AbortError') return;
            }
        }
    }
    customElements.define('suggestions-enhance', SuggestionsEnhance);
}

