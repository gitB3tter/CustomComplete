package i18n

var translationsDe = map[Key]string{
	//App-weit
	ApplicationName: "CustomComplete",

	//---
	//Geteilt
	CommonErrorService:              "Es gibt ein Problem mit dem Dienst",
	CommonAnchorNewTab:              "(öffnet in neuem Tab)",
	CommonActionCancel:              "Abbrechen",
	CommonActionSaveChanges:         "Änderungen speichern",
	CommonActionChooseLanguage:      "Sprache wählen",
	CommonActionChooseReturn:        "Antwortpräferenz wählen",
	CommonActionChooseReturnApplied: "Angewendete Antwortpräferenz",

	//Warnungen und Benachrichtigungen
	CommonNotificationInfo:          "Wichtig",
	CommonNotificationSuccess:       "Erfolg",
	CommonWarningNoUndo:             "Diese Aktion kann nicht rückgängig gemacht werden",
	CommonErrorSummaryTitle:         "Es ist ein Problem aufgetreten",
	CommonErrorMessagePrefixSR:      "Fehler",
	CommonErrorUnknown:              "Ein unbekannter Fehler ist aufgetreten",
	CommonErrorAdditionalProperties: "Unbekanntes Feld %s",

	//Wortanzahl (nur JS)
	CommonCountWordsInfo:           "Sie können bis zu %s Wörter eingeben",
	CommonCountWordsRemainingOne:   "Noch 1 Wort verfügbar",
	CommonCountWordsRemainingOther: "Noch %d Wörter verfügbar",
	CommonCountWordsToomanyOne:     "1 Wort zu viel",
	CommonCountWordsToomanyOther:   "%d Wörter zu viel",
	CommonCountWordsOne:            "1 Wort",
	CommonCountWordsOther:          "%s Wörter",

	//Anhang
	CommonAttachementMetaText: "Textdatei",
	CommonAttachementMetaJSON: "JSON",
	CommonAttachementDownload: "Herunterladen",

	//---
	//Navigation
	NavigationSkipToMainContent:        "Zum Hauptinhalt springen",
	NavigationBreadcrumb:               "Navigationspfad",
	NavigationBack:                     "Zurück",
	NavigationHome:                     "Startseite",
	NavigationTextCollection:           "Texte",
	NavigationTextCreateForm:           "Text hinzufügen",
	NavigationTextStatus:               `Status von "%s"`,
	NavigationTextEditForm:             `"%s" bearbeiten`,
	NavigationSuggestions:              "Vorschläge",
	NavigationDocumentation:            "Dokumentation",
	NavigationDocsSchemas:              "Schemata",
	NavigationDocsProblems:             "Problemtypen",
	NavigationDocsProblemGeneric:       "Allgemeines Problem",
	NavigationDocsProblemValidation:    "Validierungsproblem",
	NavigationDocsProblemInvalidParams: "Ungültige Parameter",
	NavigationDocsRels:                 "Erweiterte Link-Relationen",
	NavigationDocsDeleteForm:           "delete-form",
	NavigationDocsSpecs:                "OpenAPI-Spezifikationen",
	NavigationDocsTextsItem:            "Text - Element",
	NavigationDocsTextsCollection:      "Texte - Sammlung",

	//HTTP
	//303
	HTTPRedirectSeeOther:        "Weiterleitung zum Ergebnis",
	HTTPRedirectMetaDescription: "Sie werden zu einer anderen Seite weitergeleitet",
	HTTPRedirectAnchorLabel:     "Weiter zur nächsten Seite",

	//OpenAPI-Beschreibungen
	HTTPNotModified: "Diese Ressource wurde nicht verändert",
	HTTPVary:        "Anfrage-Header, die den Antwortinhalt beeinflussen",

	//400
	HTTPErrorRequestBodyNotAllowed: "Diese Anfrage darf keinen Anfrage-Body enthalten",
	HTTPErrorBodyNotParseable:      "Der Inhalt der Anfrage konnte nicht gelesen werden",
	//404
	HTTPErrorNotFound:            "Seite nicht gefunden",
	HTTPErrorNotFoundDescription: "Wenn Sie die Webadresse eingegeben haben, überprüfen Sie die Schreibweise. Wenn Sie die Adresse eingefügt haben, prüfen Sie, ob sie vollständig ist.",
	//406
	HTTPErrorNotAcceptable:            "Angefragtes Format nicht verfügbar",
	HTTPErrorNotAcceptableDescription: "Diese Ressource ist in keinem der angefragten Formate verfügbar.",
	HTTPErrorNotAcceptableAvailable:   "Verfügbare Formate: %s",
	//412
	HTTPErrorPrecondition:                  "Ihre Änderungen konnten nicht gespeichert werden",
	HTTPErrorPreconditionFailedDescription: "Ihre Änderungen wurden nicht gespeichert. Das aktuelle ETag stimmt nicht mit dem übergebenen ETag überein.",
	//413
	HTTPErrorContentTooLargeDescription: "Die maximale Anfragegröße beträgt %s. Reduzieren Sie die gesendete Datenmenge und versuchen Sie es erneut.",
	//415
	HTTPErrorUnsupportedMediaType:                "Medientyp nicht unterstützt",
	HTTPErrorUnsupportedMediaTypeDescription:     "Die Anfrage verwendet den Medientyp %s. Unterstützte Medientypen: %s",
	HTTPErrorUnsupportedMediaTypeMetaDescription: "Der Medientyp der gesendeten Daten wird nicht unterstützt",
	//428
	HTTPErrorPreconditionRequiredDescription: "Ihre Änderungen wurden nicht gespeichert. Die Anfrage muss das aktuelle ETag im If-Match-Header mitsenden.",
	//500
	HTTPErrorInternalServerErrorMetaDescription: "Ein interner Serverfehler ist aufgetreten.",
	HTTPErrorInternalServerErrorTryAgain:        "Versuchen Sie es später erneut.",

	//---
	//Texte
	//Felder
	//Name
	TextFieldName:              "Name",
	TextFieldNameHint:          "Zum Beispiel Faust. Maximal %d Zeichen",
	TextFieldNameRequired:      "Geben Sie einen Namen ein",
	TextFieldNameMaxCodePoints: "Der Name darf maximal %d Zeichen lang sein",
	//Datei
	TextFieldFile:                    "Datei",
	TextFieldFileHint:                "Normale Textdatei (.txt, UTF-8), maximal %d MB",
	TextFieldFileMissing:             "Wählen Sie eine Datei aus",
	TextFieldFileMinWords:            "Die ausgewählte Datei muss mindestens %d Wörter enthalten",
	TextFieldFileMaxBytes:            "Die ausgewählte Datei darf maximal %d MB groß sein",
	TextFieldFileInvalidUTF8:         "Die ausgewählte Datei muss eine UTF-8-kodierte Textdatei sein",
	TextFieldFileInvalidTXT:          "Die ausgewählte Datei muss eine Textdatei (.txt) sein",
	TextFieldFileInvalidNotParseable: "Die ausgewählte Datei konnte nicht gelesen werden. Wählen Sie die Datei erneut aus",
	//Text
	TextFieldText:            "Text",
	TextFieldTextHint:        "Geben Sie Ihren Text hier ein oder fügen Sie ihn ein. Ab %s Wörtern empfehlen wir den Datei-Upload",
	TextFieldTextHintMB:      "Maximal %d MB",
	TextFieldTextMissing:     "Geben Sie einen Text ein",
	TextFieldTextMinWords:    "Der Text muss mindestens %d Wörter enthalten",
	TextFieldTextMaxBytes:    "Der Text darf maximal %d MB groß sein",
	TextFieldTextInvalidUTF8: "Der Text muss gültiges UTF-8 sein",
	TextFieldTextInfo:        "Eingefügter Text und hochgeladene Dateien dürfen bis zu %d MB groß sein",
	TextFieldTextWordsOver:   "Datei-Upload empfohlen",

	//API-Feldvalidierung
	TextFieldNameType:              "Der Name muss eine Zeichenkette sein",
	TextFieldNameInvalidUTF8:       "Der Name enthält ungültige UTF-8-Zeichen",
	TextFieldFileNameRequired:      "Geben Sie einen Dateinamen ein",
	TextFieldFilenameMinCodePoints: "Der Dateiname muss mindestens %d Zeichen lang sein",
	TextFieldFilenameMaxBytes:      "Der Dateiname darf maximal %d Bytes lang sein",
	TextFieldFilenamePattern:       "Der Dateiname darf nur Kleinbuchstaben, Ziffern und einfache Bindestriche enthalten",
	TextFieldFilenameReserved:      "Dieser Dateiname ist vom Betriebssystem reserviert",
	TextFieldFilenameInvalid:       "Der Dateiname ist ungültig",
	TextFieldFilenameType:          "Der Dateiname muss eine Zeichenkette sein",
	TextFieldFilenameInvalidUTF8:   "Der Dateiname enthält ungültige UTF-8-Zeichen",
	TextFieldFilesizeReadOnly:      "Die Dateigröße ergibt sich aus dem Text und kann nicht direkt geändert werden",
	TextFieldFilehashReadOnly:      "Der Datei-Hash ergibt sich aus dem Text und kann nicht direkt geändert werden",

	//Sammlung
	TextCollectionMetaDescription: "Textdateien für Autovervollständigung und Wortvorschläge hochladen und verwalten",
	TextCollectionActionLabel:     "Text hinzufügen",

	//Erstellungsformular
	TextCreateFormMetaDescription: "Textdatei hochladen oder Text eingeben für Autovervollständigung und Wortvorschläge",
	TextCreateFormUnselected:      "Wählen Sie aus, wie Sie den Text bereitstellen möchten",
	TextCreateFormActionLabel:     "Text hinzufügen",
	TextCreateFormError:           "Ihr Text wurde möglicherweise nicht erstellt. Versuchen Sie es erneut oder",
	TextCreateFormErrorLabelLink:  "prüfen Sie Ihre Texte",
	TextCreateFormChoice:          "Wählen Sie eine Option",
	TextCreateFormChoiceText:      "Text eingeben oder einfügen",
	TextCreateFormChoiceFile:      "Datei hochladen",

	//Erstellen
	TextCreateFlashSuccess:      "Der Text wurde erstellt",
	TextCreateErrorChooseSource: "Geben Sie entweder 'text' oder 'file' an, nicht beides",

	//Element
	TextItemMetaDescription:     "%s. Text für Autovervollständigung und Nächstes-Wort-Vorschläge.",
	TextItemDialogDeleteKeep:    "Text behalten",
	TextItemSeeStatus:           "Status ansehen",
	TextItemGetSuggestions:      "Vorschläge erhalten",
	TextItemRefreshForWordcount: "Seite für Wortanzahl neu laden",
	TextItemManage:              "Verwalten",

	//Status
	TextStatusMetaDescription: `Verarbeitungsstatus von Text „%s"`,
	TextStatusPending:         "Verarbeitung ausstehend",
	TextStatusProcessing:      "Wird verarbeitet",
	TextStatusFailed:          "Verarbeitung fehlgeschlagen",
	TextStatusComplete:        "Verarbeitung abgeschlossen",
	TextStatusUnknown:         "Status unbekannt",
	TextStatusView:            "Status ansehen",

	//Bearbeitungsformular
	TextEditFormMetaDescription:    `Name oder Inhalt von „%s" ändern`,
	TextEditFormFileCurrent:        "Aktuelle Datei",
	TextEditFormChoice:             "Möchten Sie den Text ersetzen?",
	TextEditFormChoiceKeep:         "Bestehenden Text behalten",
	TextEditFormChoiceText:         "Neuen Text eingeben oder einfügen",
	TextEditFormChoiceFile:         "Neue Datei hochladen",
	TextEditFormError:              "Ihr Text wurde möglicherweise nicht aktualisiert. Versuchen Sie es erneut oder",
	TextEditFormErrorLabelLink:     "prüfen Sie den Text",
	TextEditFormPreconditionBefore: "Ihre Änderungen wurden nicht gespeichert. Der Text wurde nach dem Öffnen dieser Seite geändert. Prüfen Sie die",
	TextEditFormPreconditionAnchor: "aktuelle Version",
	TextEditFormPreconditionAfter:  "bevor Sie ihn überschreiben.",
	TextEditFormActionOverwrite:    "Text überschreiben",

	//Aktualisieren
	TextUpdateFlashSuccess:         "Der Text wurde aktualisiert",
	TextUpdatePreconditionRequired: "Ihre Änderungen wurden nicht gespeichert. Die Anfrage muss das aktuelle ETag im If-Match-Header mitsenden.",
	TextUpdatePreconditionFailed:   "Ihre Änderungen wurden nicht gespeichert. Das ETag des Texts und das ETag im If-Match-Header stimmen nicht überein.",

	//Löschformular
	TextDeleteFormTitle:            `„%s" löschen?`,
	TextDeleteFormMetaDescription:  `Dauerhaftes Löschen von „%s" und allen abgeleiteten Vorschlägen bestätigen`,
	TextDeleteFormWarningPermanent: ` und alle daraus abgeleiteten Vorschläge werden unwiderruflich gelöscht.`,
	TextDeleteFormActionLabel:      "Text löschen",

	//Löschen
	TextDeleteFlashSuccess:         `„%s" wurde gelöscht`,
	TextDeletePreconditionFailed:   "Der Text wurde nicht gelöscht. Rufen Sie den aktuellen Zustand ab und versuchen Sie es mit dem aktuellen ETag im If-Match-Header erneut.",
	TextDeletePreconditionRequired: "Der Text wurde nicht gelöscht. Die Anfrage muss das aktuelle ETag im If-Match-Header mitsenden.",
	TextDeletePrecondition:         "Der Text wurde nicht gelöscht, da er nach dem Öffnen der Seite geändert wurde. Prüfen Sie die aktuelle Version, bevor Sie es erneut versuchen.",
	TextDeleteNoPrecondition:       "Der Text wurde nicht gelöscht, da die Version nicht überprüft werden konnte. Prüfen Sie die aktuellen Details, bevor Sie es erneut versuchen.",

	//---
	//Verarbeitungsschritte
	ProcessingReadingText:          "Text wird gelesen",
	ProcessingCounting:             "Wörter und Wortfolgen werden gezählt",
	ProcessingPreparingSuggestions: "Vorschläge werden vorbereitet",
	ProcessingSavingSuggestions:    "Vorschläge werden gespeichert",

	//---
	//Vorschläge
	SuggestionsTitleFilteredTextID:            `Vorschläge für "%s"`,
	SuggestionsAction:                         "Vorschläge abrufen",
	SuggestionsError:                          "Vorschläge konnten nicht geladen werden",
	SuggestionsFilterTextError:                "Filtern nach dem ausgewählten Text nicht möglich",
	SuggestionsFilterTextNotFound:             "Der ausgewählte Text wurde nicht gefunden",
	SuggestionsFilterTextProcessingFailed:     "Die Verarbeitung des ausgewählten Texts ist fehlgeschlagen",
	SuggestionsFilterTextProcessingIncomplete: "Die Verarbeitung des ausgewählten Texts ist noch nicht abgeschlossen",
	SuggestionsBackToLast:                     "Zurück zu den letzten Vorschlägen",
	SuggestionsNoTexts:                        "Fügen Sie zuerst einen Text hinzu, um Vorschläge zu erhalten",

	//---
	//Startseite
	HomeSubtitle:                 "Texte hochladen für Autovervollständigung und Nächstes-Wort-Vorschläge",
	HomeMetaDescription:          "Textkorpora hochladen und N-Gramm-basierte Autovervollständigung sowie Wortvorschläge abrufen",
	HomeNavHeading:               "Bereiche",
	HomeTextsDescription:         "Texte hochladen, einsehen und verwalten",
	HomeSuggestionsDescription:   "Autovervollständigung und Nächstes-Wort-Vorschläge abrufen",
	HomeDocumentationDescription: "API-Spezifikationen und Dokumentation",

	//---
	//Dokumentation
	DocsTitlePrefix:                   "Dokumentation",
	DocsFiles:                         "Dateien",
	DocsRelsCollectionMetaDescription: "Vom Dienst verwendete benutzerdefinierte Link-Relationen",
	DocsRelsDeleteFormBodyDescription: "Die Zielressource stellt ein Formular zum dauerhaften Löschen der Kontextressource bereit.",
	DocsRelsDeleteFormMetaDescription: "Definiert den erweiterten Link-Relationstyp delete-form",

	//Problemtypen
	DocsProblemValidationDescription:         "Die Anfrage hat die Validierung nicht bestanden. Details zu den einzelnen Feldern finden Sie unter errors.",
	DocsProblemTitleDescription:              "Kurze Zusammenfassung des Problemtyps. Inhaltlich stets identisch, lediglich lokalisiert.",
	DocsProblemInvalidParamsDescription:      "Liste aller fehlerhaften Felder mit stabilem Fehlercode, Parametern, Feldname und lokalisierter Begründung.",
	DocsProblemInvalidParamNameDescription:   "Name des fehlerhaften Felds oder Parameters",
	DocsProblemInvalidParamCodeDescription:   "Stabiler maschinenlesbarer Code der verletzten Bedingung",
	DocsProblemInvalidParamParamsDescription: "Zusätzliche Parameter zur Spezifikation der Bedingung, etwa Grenzwerte oder erwartete Typen",
	DocsProblemInvalidParamReasonDescription: "Menschenlesbare, lokalisierte Fehlermeldung",

	//Spezifikationsbeispiele
	DocsSpecsExampleTextName:     "Beispieltext",
	DocsSpecsExampleTextFilename: "beispiel-dateiname",
	DocsSpecsExampleLangTag:      "de",
	DocsSpecsExampleOtherLangTag: "en",
}
