package i18n

var translationsEn = map[Key]string{
	//App-wide
	ApplicationName: "CustomComplete",

	//---
	//Shared
	//eg. for (POST) fetch throwing, unsure if succeeded
	CommonErrorService:              "There is a problem with the service",
	CommonAnchorNewTab:              "(opens in a new tab)",
	CommonActionCancel:              "Cancel",
	CommonActionSaveChanges:         "Save changes",
	CommonActionChooseLanguage:      "Choose Language",
	CommonActionChooseReturn:        "Choose response format",
	CommonActionChooseReturnApplied: "format choice applied",

	//Warnings and Notification
	CommonNotificationInfo:          "Important",
	CommonNotificationSuccess:       "Success",
	CommonWarningNoUndo:             "You cannot undo this action",
	CommonErrorSummaryTitle:         "There is a problem",
	CommonErrorMessagePrefixSR:      "Error",
	CommonErrorUnknown:              "An unknown error occured",
	CommonErrorAdditionalProperties: "unknown field %s",

	// counting words
	CommonCountWordsInfo:           "You can enter up to %s words",
	CommonCountWordsRemainingOne:   "You have 1 word remaining",
	CommonCountWordsRemainingOther: "You have %d words remaining",
	CommonCountWordsToomanyOne:     "You have 1 word too many",
	CommonCountWordsToomanyOther:   "You have %d words too many",
	CommonCountWordsOne:            "1 word",
	CommonCountWordsOther:          "%s words",
	//attachment
	CommonAttachementMetaText: "Plain Text",
	CommonAttachementMetaJSON: "JSON",
	CommonAttachementDownload: "Download",

	//---
	//Navigation
	//General
	NavigationSkipToMainContent:        "Skip to main content",
	NavigationBreadcrumb:               "Breadcrumbs",
	NavigationBack:                     "Back",
	NavigationHome:                     "Home",
	NavigationTextCollection:           "Texts",
	NavigationTextCreateForm:           "Add Text",
	NavigationTextStatus:               "Status of \"%s\"",
	NavigationTextEditForm:             `Edit "%s"`,
	NavigationSuggestions:              "Suggestions",
	NavigationDocumentation:            "Documentation",
	NavigationDocsSchemas:              "Schemas",
	NavigationDocsProblems:             "Problems",
	NavigationDocsProblemGeneric:       "Generic Problem",
	NavigationDocsProblemValidation:    "Validation Problem",
	NavigationDocsProblemInvalidParams: "Invalid Parameters",
	NavigationDocsRels:                 "Extension Link Relations",
	NavigationDocsDeleteForm:           "delete-form",
	NavigationDocsSpecs:                "OpenAPI Specifications",
	NavigationDocsTextsItem:            "Text - Item",
	NavigationDocsTextsCollection:      "Texts - Collection",

	//---
	//HTTP
	//303
	HTTPRedirectSeeOther:        "You are being redirected",
	HTTPRedirectMetaDescription: "You are being redirected to another page",
	HTTPRedirectAnchorLabel:     "Continue to the next page",

	//OpenAPI descriptions
	HTTPNotModified: "This resource has not changed",
	HTTPVary:        "Requests headers influencing response content",

	//Errors
	//400
	HTTPErrorRequestBodyNotAllowed: "This request must not include a body",
	HTTPErrorBodyNotParseable:      "The request contents can not be read",
	//404
	HTTPErrorNotFound:            "Page not found",
	HTTPErrorNotFoundDescription: "If you typed the web address, check it is correct. If you pasted the web address, check you copied the entire address.",
	//406
	HTTPErrorNotAcceptable:            "Requested format not available",
	HTTPErrorNotAcceptableDescription: "This page is not available in any of the formats you asked for.",
	HTTPErrorNotAcceptableAvailable:   "Available formats are: %s",
	//412
	HTTPErrorPrecondition:                  "Your changes could not be saved",
	HTTPErrorPreconditionFailedDescription: "Your changes have not been saved. The current ETag and given ETag do not match",
	//413
	HTTPErrorContentTooLargeDescription: "Max Request size is %s. Reduce the amount of data you send and try again.",
	//415
	HTTPErrorUnsupportedMediaType:            "Content type not supported",
	HTTPErrorUnsupportedMediaTypeDescription: "The request uses the content type %s. These content types are supported: %s",
	//requires sent mediatypes.Type.String() and processable mediatypes.Types, contains HTML
	HTTPErrorUnsupportedMediaTypeMetaDescription: "The requested content type is not supported",
	//428
	HTTPErrorPreconditionRequiredDescription: "Your changes have not been saved. The request must be conditional using If-Match with the current ETag",
	//500
	HTTPErrorInternalServerErrorMetaDescription: "The server could not complete your request",
	HTTPErrorInternalServerErrorTryAgain:        "Try again later",

	//---
	//Texts
	//Fields
	//Name
	TextFieldName:              "Name",
	TextFieldNameHint:          "For example Hamlet. Must be %d characters or fewer",
	TextFieldNameRequired:      "Enter a name",
	TextFieldNameMaxCodePoints: "Name must be %d characters or fewer",
	//File
	TextFieldFile:                    "File",
	TextFieldFileHint:                "Must be a .txt file, %d MB or smaller",
	TextFieldFileMissing:             "Select a file",
	TextFieldFileMinWords:            "The selected file must contain at least %d words",
	TextFieldFileMaxBytes:            "The selected file must be %d MB or smaller",
	TextFieldFileInvalidUTF8:         "The selected file must be a UTF-8 plain text file",
	TextFieldFileInvalidTXT:          "The selected file must be a text file (.txt)",
	TextFieldFileInvalidNotParseable: "The selected file could not be read. Try choosing the file again",
	//Text
	TextFieldText:            "Text",
	TextFieldTextHint:        "Type or paste your text here. For texts with more than %s words uploading a file is recommended",
	TextFieldTextHintMB:      "Maximum size is %d MB",
	TextFieldTextMissing:     "Enter text",
	TextFieldTextMinWords:    "Text must be at least %d words",
	TextFieldTextMaxBytes:    "Text must be %d MB or smaller",
	TextFieldTextInvalidUTF8: "Text must be valid UTF-8",
	TextFieldTextInfo:        "Pasted text and uploaded files can be up to %d MB",
	TextFieldTextWordsOver:   "file upload recommended",

	//API field validation
	TextFieldNameType:              "Name must be a string",
	TextFieldNameInvalidUTF8:       "Name contains invalid UTF-8 characters",
	TextFieldFileNameRequired:      "Enter a filename",
	TextFieldFilenameMinCodePoints: "Filename must be at least %d characters",
	TextFieldFilenameMaxBytes:      "Filename must be %d bytes or fewer",
	TextFieldFilenamePattern:       "Filename may only contain lowercase letters, numbers, and single hyphens",
	TextFieldFilenameReserved:      "Filename is reserved by the operating system",
	TextFieldFilenameInvalid:       "Filename is invalid",
	TextFieldFilenameType:          "Filename must be a string",
	TextFieldFilenameInvalidUTF8:   "Filename contains invalid UTF-8 characters",
	TextFieldFilesizeReadOnly:      "Filesize is derived from the text and cannot be changed directly",
	TextFieldFilehashReadOnly:      "Filehash is derived from the text and cannot be changed directly",

	//collection
	TextCollectionMetaDescription: "Upload and manage text files for autocomplete and next word suggestions",
	TextCollectionActionLabel:     "Add a text",

	//create-form
	TextCreateFormMetaDescription: "Upload a text file to use for autocomplete and next word suggestions",
	TextCreateFormUnselected:      "Choose an option",
	TextCreateFormActionLabel:     "Add text",
	TextCreateFormError:           "Your text may not have been created. Try again, or",
	TextCreateFormErrorLabelLink:  "check your texts",
	TextCreateFormChoice:          "How do you want to provide the text?",
	TextCreateFormChoiceText:      "Type or paste text",
	TextCreateFormChoiceFile:      "Upload a file",

	//create
	TextCreateFlashSuccess:      "The text has been created",
	TextCreateErrorChooseSource: "Provide either 'text' or 'file', but not both",

	//item
	TextItemMetaDescription:     "%s. Text used for autocomplete and next word suggestions.",
	TextItemDialogDeleteKeep:    "Keep text",
	TextItemSeeStatus:           "View Status",
	TextItemGetSuggestions:      "Get Suggestions",
	TextItemRefreshForWordcount: "Refresh page for word count",
	TextItemManage:              "Manage",

	//status
	TextStatusMetaDescription: `Processing status of text "%s"`,
	TextStatusPending:         "Waiting to be processed",
	TextStatusProcessing:      "Processing",
	TextStatusFailed:          "Processing failed",
	TextStatusComplete:        "Processing complete",
	TextStatusUnknown:         "Status unknown",
	TextStatusView:            "View status",

	//edit-form
	//requires Text.Name string
	TextEditFormMetaDescription:    `Change the name or file for "%s"`,
	TextEditFormFileCurrent:        "Current file",
	TextEditFormChoice:             "Do you want to replace the text?",
	TextEditFormChoiceKeep:         "Keep text",
	TextEditFormChoiceText:         "Type or paste new text",
	TextEditFormChoiceFile:         "Upload a replacement file",
	TextEditFormError:              "Your text might not have been updated. Try again, or",
	TextEditFormErrorLabelLink:     "check your text",
	TextEditFormPreconditionBefore: "Your changes have not been saved. The text was updated after you opened this page. Check the",
	TextEditFormPreconditionAnchor: "latest version",
	TextEditFormPreconditionAfter:  "before overwriting",
	TextEditFormActionOverwrite:    "Overwrite text",

	//update
	TextUpdateFlashSuccess:         "The text has been updated",
	TextUpdatePreconditionRequired: "Your changes have not been saved. The request must be conditional using If-Match with current ETag",
	TextUpdatePreconditionFailed:   "Your changes have not been saved. The texts ETag and ETag given in If-Match do not match.",

	//delete-form
	TextDeleteFormTitle:            `Delete "%s"?`,
	TextDeleteFormMetaDescription:  `Confirm the permanent deletion of "%s" and all derived suggestions`,
	TextDeleteFormWarningPermanent: ` and all derived suggestions will be permanently deleted`,
	TextDeleteFormActionLabel:      "Delete text",

	//delete
	TextDeleteFlashSuccess:         `"%s" has been deleted`,
	TextDeletePreconditionFailed:   "The text was not deleted. Fetch the latest state and retry with the current ETag in the If-Match header.",
	TextDeletePreconditionRequired: "The text was not deleted. The request must be conditional using If-Match with the current ETag.",
	TextDeletePrecondition:         "The text was not deleted because it was modified after you opened the page. Check the latest version before trying again.",
	TextDeleteNoPrecondition:       "The text was not deleted because its version could not be verified. Review its current details before trying again.",

	//---
	//Processing
	//stages
	ProcessingReadingText:          "Reading text",
	ProcessingCounting:             "Counting words and phrases",
	ProcessingPreparingSuggestions: "Preparing suggestions",
	ProcessingSavingSuggestions:    "Saving suggestions",

	//---
	//suggestions
	SuggestionsTitleFilteredTextID:            `Suggestions for "%s"`,
	SuggestionsAction:                         "Get Suggestions",
	SuggestionsError:                          "Suggestions could not be loaded",
	SuggestionsFilterTextError:                "Cannot filter by the selected text",
	SuggestionsFilterTextNotFound:             "The selected text was not found",
	SuggestionsFilterTextProcessingFailed:     "Processing of the selected text failed",
	SuggestionsFilterTextProcessingIncomplete: "Processing of the selected text is not yet complete",
	SuggestionsBackToLast:                     "Return to previous suggestions",
	SuggestionsNoTexts:                        "Add a text first to get suggestions",

	//---
	//Home
	HomeSubtitle:                 "Upload texts to get autocomplete and next word suggestions",
	HomeMetaDescription:          "Upload text corpora and get n-gram autocomplete and next-word suggestions",
	HomeNavHeading:               "Sections",
	HomeTextsDescription:         "Upload, inspect and manage texts",
	HomeSuggestionsDescription:   "Get autocomplete and next word suggestions",
	HomeDocumentationDescription: "API specifications and documentation",

	//---
	//Docs
	DocsTitlePrefix:                   "Documentation",
	DocsFiles:                         "Files",
	DocsRelsCollectionMetaDescription: "Custom extension link relation types used by the service",
	DocsRelsDeleteFormBodyDescription: "The target resource contains a form for permanently deleting the context resource",
	DocsRelsDeleteFormMetaDescription: "Defines the delete-form link relation type",

	//Problems
	DocsProblemValidationDescription:         "The request did not pass validation. See errors for field-level details.",
	DocsProblemTitleDescription:              "Short summary of the problem type. Constant across occurrences, localized via content negotiation.",
	DocsProblemInvalidParamsDescription:      "List of field-level validation errors with a stable code, parameters, field name, and localized reason.",
	DocsProblemInvalidParamNameDescription:   "Name of the invalid field or parameter",
	DocsProblemInvalidParamCodeDescription:   "Stable machine-readable code identifying the violated constraint",
	DocsProblemInvalidParamParamsDescription: "Additional constraint parameters such as limits or expected types",
	DocsProblemInvalidParamReasonDescription: "Human-readable localized explanation of the error",

	//specifications
	DocsSpecsExampleTextName:     "ExampleText",
	DocsSpecsExampleTextFilename: "example-filename",
	DocsSpecsExampleLangTag:      "en",
	DocsSpecsExampleOtherLangTag: "de",
}
