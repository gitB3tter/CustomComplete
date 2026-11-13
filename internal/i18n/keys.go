package i18n

const (
	//App-wide
	ApplicationName Key = "ApplicationName"

	//---
	//Shared
	//eg. for (POST) fetch throwing, unsure if succeeded
	CommonErrorService Key = "CommonErrorService"
	//same url href anchor label
	CommonAnchorNewTab Key = "CommonAnchorNewTab"
	//Actions
	CommonActionCancel              Key = "CommonActionCancel"
	CommonActionSaveChanges         Key = "CommonActionSaveChanges"
	CommonActionChooseLanguage      Key = "CommonActionChooseLanguage"
	CommonActionChooseReturn        Key = "CommonActionChooseReturn"
	CommonActionChooseReturnApplied Key = "CommonActionChooseReturnApplied"
	//Warnings and Notification
	CommonNotificationInfo     Key = "CommonNotificationInfo"
	CommonNotificationSuccess  Key = "CommonNotificationSuccess"
	CommonWarningNoUndo        Key = "CommonWarningNoUndo"
	CommonErrorSummaryTitle    Key = "CommonErrorSummaryTitle"
	CommonErrorMessagePrefixSR Key = "CommonErrorMessagePrefixSR"
	CommonErrorUnknown         Key = "CommonErrorUnknown"
	//requires fieldname string
	CommonErrorAdditionalProperties Key = "CommonErrorAdditionalProperties"

	//counting words
	//requires language aware formatted max words strings
	CommonCountWordsInfo         Key = "CommonCountWordsInfo"
	CommonCountWordsRemainingOne Key = "CommonCountWordsRemainingOne"
	//requires remaining words count int
	CommonCountWordsRemainingOther Key = "CommonCountWordsRemainingOther"
	CommonCountWordsToomanyOne     Key = "CommonCountWordsToomanyOne"
	//requires exceeding words count int
	CommonCountWordsToomanyOther Key = "CommonCountWordsToomanyOther"
	//advisory count only, requires field-specific phrasing for count exceeding
	//requires language aware formatted max words strings
	CommonCountWordsOne Key = "CommonCountWordsOne"
	//requires language aware formatted max words strings
	CommonCountWordsOther Key = "CommonCountWordsOther"
	//attachment
	CommonAttachementMetaText Key = "CommonAttachementMetaText"
	CommonAttachementMetaJSON Key = "CommonAttachementMetaJSON"
	CommonAttachementDownload Key = "CommonAttachementDownload"

	//---
	//Navigation
	//General
	NavigationSkipToMainContent Key = "NavigationSkipToMainContent"
	NavigationBreadcrumb        Key = "NavigationBreadcrumb"
	NavigationBack              Key = "NavigationBack"
	//Pagetitles
	NavigationHome           Key = "NavigationHome"
	NavigationTextCollection Key = "NavigationTextCollection"
	NavigationTextCreateForm Key = "NavigationTextCreateForm"
	//(textitem title === name)
	//requires Text.Name string
	NavigationTextStatus Key = "NavigationTextStatus"
	//requires Text.Name string
	NavigationTextEditForm Key = "NavigationTextEditForm"
	//requires Text.Name string
	NavigationTextDeleteForm           Key = "NavigationTextDeleteForm"
	NavigationSuggestions              Key = "NavigationSuggestions"
	NavigationDocumentation            Key = "NavigationDocumentation"
	NavigationDocsSchemas              Key = "NavigationDocsSchemas"
	NavigationDocsProblems             Key = "NavigationDocsProblems"
	NavigationDocsProblemGeneric       Key = "NavigationDocsProblemGeneric"
	NavigationDocsProblemValidation    Key = "NavigationDocsProblemValidation"
	NavigationDocsProblemInvalidParams Key = "NavigationDocsProblemInvalidParams"
	NavigationDocsRels                 Key = "NavigationDocsRels"
	NavigationDocsDeleteForm           Key = "NavigationDocsDeleteForm"
	NavigationDocsSpecs                Key = "NavigationDocsSpecs"
	NavigationDocsTextsItem            Key = "NavigationDocsTextsItem"
	NavigationDocsTextsCollection      Key = "NavigationDocsTextsCollection"

	//---
	//HTTP
	//Redirects
	//303
	HTTPRedirectSeeOther        Key = "HTTPRedirectSeeOther"
	HTTPRedirectMetaDescription Key = "HTTPRedirectMetaDescription"
	HTTPRedirectAnchorLabel     Key = "HTTPRedirectAnchorLabel"

	//OpenAPI description
	HTTPNotModified Key = "HTTPNotModified"
	//OpenAPI description
	HTTPVary Key = "HTTPVary"

	//Errors
	//400
	HTTPErrorRequestBodyNotAllowed Key = "HTTPErrorRequestBodyNotAllowed"
	HTTPErrorBodyNotParseable      Key = "HTTPErrorBodyNotParseable"
	//404
	HTTPErrorNotFound            Key = "HTTPErrorNotFound"
	HTTPErrorNotFoundDescription Key = "HTTPErrorNotFoundDescription"
	//406
	HTTPErrorNotAcceptable            Key = "HTTPErrorNotAcceptable"
	HTTPErrorNotAcceptableDescription Key = "HTTPErrorNotAcceptableDescription"
	//requires supported mediatypes.Types
	HTTPErrorNotAcceptableAvailable Key = "HTTPErrorNotAcceptableAvailable"
	//412
	HTTPErrorPrecondition                  Key = "HTTPErrorPrecondition"
	HTTPErrorPreconditionFailedDescription Key = "HTTPErrorPreconditionFailedDescription"
	//413
	//requires max content size fmt string
	HTTPErrorContentTooLargeDescription Key = "HTTPErrorContentTooLargeDescription"
	//415
	HTTPErrorUnsupportedMediaType                Key = "HTTPErrorUnsupportedMediaType"
	HTTPErrorUnsupportedMediaTypeMetaDescription Key = "HTTPErrorUnsupportedMediaTypeMetaDescription"
	//requires provided mediatypes.Type, supported mediatypes.Types
	HTTPErrorUnsupportedMediaTypeDescription Key = "HTTPErrorUnsupportedMediaTypeDescription"
	//428
	HTTPErrorPreconditionRequiredDescription Key = "HTTPErrorPreconditionRequiredDescription"
	//500
	HTTPErrorInternalServerErrorMetaDescription Key = "HTTPErrorInternalServerErrorMetaDescription"
	HTTPErrorInternalServerErrorTryAgain        Key = "HTTPErrorInternalServerErrorTryAgain"

	//---
	//Texts
	//Fields
	//Name
	TextFieldName Key = "TextFieldName"
	//requires maxCodePoints int
	TextFieldNameHint     Key = "TextFieldNameHint"
	TextFieldNameRequired Key = "TextFieldNameRequired"
	//requires maxCodePoints int
	TextFieldNameMaxCodePoints Key = "TextFieldNameMaxCodePoints"
	//File
	TextFieldFile Key = "TextFieldFile"
	//requires maxBytes in MB int
	TextFieldFileHint    Key = "TextFieldFileHint"
	TextFieldFileMissing Key = "TextFieldFileMissing"
	//requires minWords int
	TextFieldFileMinWords Key = "TextFieldFileMinWords"
	//requires maxBytes in MB int
	TextFieldFileMaxBytes            Key = "TextFieldFileMaxBytes"
	TextFieldFileInvalidNotParseable Key = "TextFieldFileInvalidNotParseable"
	TextFieldFileInvalidUTF8         Key = "TextFieldFileInvalidUTF8"
	TextFieldFileInvalidTXT          Key = "TextFieldFileInvalidTXT"
	//Text
	TextFieldText Key = "TextFieldText"
	//requires wordlimit max as language-aware (decimal separator) formatted string
	TextFieldTextHint Key = "TextFieldTextHint"
	//requires maxBytes in MB int, advisory text no-js only (gets hidden in favor of wordcount)
	TextFieldTextHintMB  Key = "TextFieldTextHintMB"
	TextFieldTextMissing Key = "TextFieldTextMissing"
	//requires wordlimit min int
	TextFieldTextMinWords Key = "TextFieldTextMinWords"
	//requires maxBytes in MB int
	TextFieldTextMaxBytes    Key = "TextFieldTextMaxBytes"
	TextFieldTextInvalidUTF8 Key = "TextFieldTextInvalidUTF8"
	//advise to upload file instead when typing more than max words
	//requires maxBytes in MB int
	TextFieldTextInfo Key = "TextFieldTextInfo"
	//advisory js only
	TextFieldTextWordsOver Key = "TextFieldTextWordsOver"

	//api
	//Name
	TextFieldNameType        Key = "TextFieldNameType"
	TextFieldNameInvalidUTF8 Key = "TextFieldNameInvalidUTF8"
	//Filename
	TextFieldFileNameRequired Key = "TextFieldFileNameRequired"
	//requires minCodePoints int
	TextFieldFilenameMinCodePoints Key = "TextFieldFilenameMinCodePoints"
	//requires maxBytes int
	TextFieldFilenameMaxBytes    Key = "TextFieldFilenameMaxBytes"
	TextFieldFilenamePattern     Key = "TextFieldFilenamePattern"
	TextFieldFilenameInvalid     Key = "TextFieldFilenameInvalid"
	TextFieldFilenameType        Key = "TextFieldFilenameType"
	TextFieldFilenameInvalidUTF8 Key = "TextFieldFilenameInvalidUTF8"
	//readonly
	TextFieldFilesizeReadOnly Key = "TextFieldFilesizeReadOnly"
	TextFieldFilehashReadOnly Key = "TextFieldFilehashReadOnly"

	//collection
	TextCollectionMetaDescription Key = "TextCollectionMetaDescription"
	TextCollectionActionLabel     Key = "TextCollectionActionLabel"

	//create-form
	TextCreateFormMetaDescription Key = "TextCreateFormMetaDescription"
	TextCreateFormChoice          Key = "TextCreateFormChoice"
	TextCreateFormChoiceText      Key = "TextCreateFormChoiceText"
	TextCreateFormChoiceFile      Key = "TextCreateFormChoiceFile"
	TextCreateFormUnselected      Key = "TextCreateFormUnselected"
	TextCreateFormActionLabel     Key = "TextCreateFormActionLabel"
	//fetch threw on POST, outcome unknown, append anchor with TextCreateFormErrorLabelLink
	TextCreateFormError Key = "TextCreateFormError"
	//link to Collection, tell user to check texts
	TextCreateFormErrorLabelLink Key = "TextCreateFormErrorLabelLink"

	//create
	TextCreateFlashSuccess Key = "TextCreateFlashSuccess"

	//item
	//requires Text.Name string
	TextItemMetaDescription     Key = "TextItemMetaDescription"
	TextItemDialogDeleteKeep    Key = "TextItemDialogDeleteKeep"
	TextItemSeeStatus           Key = "TextItemSeeStatus"
	TextItemGetSuggestions      Key = "TextItemGetSuggestions"
	TextItemRefreshForWordcount Key = "TextItemRefreshForWordcount"
	TextItemManage              Key = "TextItemManage"

	//status
	//requires Text.Name string
	TextStatusMetaDescription Key = "TextStatusMetaDescription"
	TextStatusPending         Key = "TextStatusPending"
	TextStatusProcessing      Key = "TextStatusProcessing"
	TextStatusFailed          Key = "TextStatusFailed"
	TextStatusComplete        Key = "TextStatusComplete"
	TextStatusUnknown         Key = "TextStatusUnknown"
	TextStatusView            Key = "TextStatusView"

	//edit-form
	//requires Text.Name string
	TextEditFormMetaDescription Key = "TextEditFormMetaDescription"
	TextEditFormChoice          Key = "TextEditFormChoice"
	TextEditFormChoiceKeep      Key = "TextEditFormChoiceKeep"
	TextEditFormChoiceText      Key = "TextEditFormChoiceText"
	TextEditFormChoiceFile      Key = "TextEditFormChoiceFile"
	TextEditFormFileCurrent     Key = "TextEditFormFileCurrent"
	//fetch threw on POST, outcome unknown, append anchor with TextCreateFormErrorLabelLink
	TextEditFormError Key = "TextEditFormError"
	//link to Item, tell user to check item, requires Text.ID int64
	TextEditFormErrorLabelLink Key = "TextEditFormErrorLabelLink"
	//precondition/conflict
	TextEditFormPreconditionBefore Key = "TextEditFormPreconditionBefore"
	TextEditFormPreconditionAnchor Key = "TextEditFormPreconditionAnchor"
	TextEditFormPreconditionAfter  Key = "TextEditFormPreconditionAfter"
	TextEditFormActionOverwrite    Key = "TextEditFormActionOverwrite"

	//update
	TextUpdateFlashSuccess Key = "TextUpdateFlashSuccess"
	//api-client
	TextUpdatePreconditionRequired Key = "TextUpdatePreconditionRequired"
	//api-client
	TextUpdatePreconditionFailed Key = "TextUpdatePreconditionFailed"

	//delete-form
	//requires Text.Name string
	TextDeleteFormMetaDescription Key = "TextDeleteFormMetaDescription"
	//requires Text.Name string prended
	TextDeleteFormWarningPermanent Key = "TextDeleteFormWarningPermanent"
	TextDeleteFormActionLabel      Key = "TextDeleteFormActionLabel"
	//requires Text.Name string

	//delete
	//requires Text.Name string
	TextDeleteFlashSuccess Key = "TextDeleteFlashSuccess"
	//api-client
	TextDeletePreconditionFailed Key = "TextDeletePreconditionFailed"
	//api-client
	TextDeletePreconditionRequired Key = "TextDeletePreconditionRequired"
	//html
	TextDeletePrecondition Key = "TextDeletePrecondition"
	//html, shouldnt happen
	TextDeleteNoPrecondition Key = "TextDeleteNoPrecondition"

	//---
	//Processing
	//stages
	ProcessingReadingText          Key = "ProcessingReadingText"
	ProcessingCounting             Key = "ProcessingCounting"
	ProcessingPreparingSuggestions Key = "ProcessingPreparingSuggestions"
	ProcessingSavingSuggestions    Key = "ProcessingSavingSuggestions"

	//---
	//suggestions
	//requires textID int(64)
	SuggestionsTitleFilteredTextID            Key = "SuggestionsTitleFilteredTextID"
	SuggestionsAction                         Key = "SuggestionsAction"
	SuggestionsError                          Key = "SuggestionsError"
	SuggestionsFilterTextError                Key = "SuggestionsFilterTextError"
	SuggestionsFilterTextNotFound             Key = "SuggestionsFilterTextNotFound"
	SuggestionsFilterTextProcessingFailed     Key = "SuggestionsFilterTextProcessingFailed"
	SuggestionsFilterTextProcessingIncomplete Key = "SuggestionsFilterTextProcessingIncomplete"
	SuggestionsBackToLast                     Key = "SuggestionsBackToLast"
	SuggestionsNoTexts                        Key = "SuggestionsNoTexts"

	//---
	//Home
	HomeSubtitle                 Key = "HomeSubtitle"
	HomeMetaDescription          Key = "HomeMetaDescription"
	HomeNavHeading               Key = "HomeNavHeading"
	HomeTextsDescription         Key = "HomeTextsDescription"
	HomeSuggestionsDescription   Key = "HomeSuggestionsDescription"
	HomeDocumentationDescription Key = "HomeDocumentationDescription"

	//---
	//Docs
	DocsFiles Key = "DocsFiles"
	//rels
	DocsRelsCollectionMetaDescription Key = "DocsRelsCollectionMetaDescription"
	DocsRelsDeleteFormBodyDescription Key = "DocsRelsDeleteFormBodyDescription"
	DocsRelsDeleteFormMetaDescription Key = "DocsRelsDeleteFormMetaDescription"

	//Problems
	DocsProblemValidationDescription    Key = "DocsProblemValidationDescription"
	DocsProblemTitleDescription         Key = "DocsProblemTitleDescription"
	DocsProblemInvalidParamsTitle       Key = "DocsProblemInvalidParamsTitle"
	DocsProblemInvalidParamsDescription Key = "DocsProblemInvalidParamsDescription"
	//extensions
	DocsProblemInvalidParamNameTitle         Key = "DocsProblemInvalidParamNameTitle"
	DocsProblemInvalidParamNameDescription   Key = "DocsProblemInvalidParamNameDescription"
	DocsProblemInvalidParamCodeTitle         Key = "DocsProblemInvalidParamCodeTitle"
	DocsProblemInvalidParamCodeDescription   Key = "DocsProblemInvalidParamCodeDescription"
	DocsProblemInvalidParamParamsTitle       Key = "DocsProblemInvalidParamParamsTitle"
	DocsProblemInvalidParamParamsDescription Key = "DocsProblemInvalidParamParamsDescription"
	DocsProblemInvalidParamReasonTitle       Key = "DocsProblemInvalidParamReasonTitle"
	DocsProblemInvalidParamReasonDescription Key = "DocsProblemInvalidParamReasonDescription"

	//specifications
	DocsSpecsExampleTextName     Key = "DocsSpecsExampleTextName"
	DocsSpecsExampleTextFilename Key = "DocsSpecsExampleTextFilename"
	DocsSpecsExampleLangTag      Key = "DocsSpecsExampleLangTag"
	DocsSpecsExampleOtherLangTag Key = "DocsSpecsExampleOtherLangTag"
)
