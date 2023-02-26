// Code generated by "stringer -type=APIErrorCode -trimprefix=Err api-errors.go"; DO NOT EDIT.

package cmd

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ErrNone-0]
	_ = x[ErrAccessDenied-1]
	_ = x[ErrBadDigest-2]
	_ = x[ErrEntityTooSmall-3]
	_ = x[ErrEntityTooLarge-4]
	_ = x[ErrPolicyTooLarge-5]
	_ = x[ErrIncompleteBody-6]
	_ = x[ErrInternalError-7]
	_ = x[ErrInvalidAccessKeyID-8]
	_ = x[ErrAccessKeyDisabled-9]
	_ = x[ErrInvalidBucketName-10]
	_ = x[ErrInvalidDigest-11]
	_ = x[ErrInvalidRange-12]
	_ = x[ErrInvalidRangePartNumber-13]
	_ = x[ErrInvalidCopyPartRange-14]
	_ = x[ErrInvalidCopyPartRangeSource-15]
	_ = x[ErrInvalidMaxKeys-16]
	_ = x[ErrInvalidEncodingMethod-17]
	_ = x[ErrInvalidMaxUploads-18]
	_ = x[ErrInvalidMaxParts-19]
	_ = x[ErrInvalidPartNumberMarker-20]
	_ = x[ErrInvalidPartNumber-21]
	_ = x[ErrInvalidRequestBody-22]
	_ = x[ErrInvalidCopySource-23]
	_ = x[ErrInvalidMetadataDirective-24]
	_ = x[ErrInvalidCopyDest-25]
	_ = x[ErrInvalidPolicyDocument-26]
	_ = x[ErrInvalidObjectState-27]
	_ = x[ErrMalformedXML-28]
	_ = x[ErrMissingContentLength-29]
	_ = x[ErrMissingContentMD5-30]
	_ = x[ErrMissingRequestBodyError-31]
	_ = x[ErrMissingSecurityHeader-32]
	_ = x[ErrNoSuchBucket-33]
	_ = x[ErrNoSuchBucketPolicy-34]
	_ = x[ErrNoSuchBucketLifecycle-35]
	_ = x[ErrNoSuchLifecycleConfiguration-36]
	_ = x[ErrInvalidLifecycleWithObjectLock-37]
	_ = x[ErrNoSuchBucketSSEConfig-38]
	_ = x[ErrNoSuchCORSConfiguration-39]
	_ = x[ErrNoSuchWebsiteConfiguration-40]
	_ = x[ErrReplicationConfigurationNotFoundError-41]
	_ = x[ErrRemoteDestinationNotFoundError-42]
	_ = x[ErrReplicationDestinationMissingLock-43]
	_ = x[ErrRemoteTargetNotFoundError-44]
	_ = x[ErrReplicationRemoteConnectionError-45]
	_ = x[ErrReplicationBandwidthLimitError-46]
	_ = x[ErrBucketRemoteIdenticalToSource-47]
	_ = x[ErrBucketRemoteAlreadyExists-48]
	_ = x[ErrBucketRemoteLabelInUse-49]
	_ = x[ErrBucketRemoteArnTypeInvalid-50]
	_ = x[ErrBucketRemoteArnInvalid-51]
	_ = x[ErrBucketRemoteRemoveDisallowed-52]
	_ = x[ErrRemoteTargetNotVersionedError-53]
	_ = x[ErrReplicationSourceNotVersionedError-54]
	_ = x[ErrReplicationNeedsVersioningError-55]
	_ = x[ErrReplicationBucketNeedsVersioningError-56]
	_ = x[ErrReplicationDenyEditError-57]
	_ = x[ErrRemoteTargetDenyEditError-58]
	_ = x[ErrReplicationNoExistingObjects-59]
	_ = x[ErrObjectRestoreAlreadyInProgress-60]
	_ = x[ErrNoSuchKey-61]
	_ = x[ErrNoSuchUpload-62]
	_ = x[ErrInvalidVersionID-63]
	_ = x[ErrNoSuchVersion-64]
	_ = x[ErrNotImplemented-65]
	_ = x[ErrPreconditionFailed-66]
	_ = x[ErrRequestTimeTooSkewed-67]
	_ = x[ErrSignatureDoesNotMatch-68]
	_ = x[ErrMethodNotAllowed-69]
	_ = x[ErrInvalidPart-70]
	_ = x[ErrInvalidPartOrder-71]
	_ = x[ErrAuthorizationHeaderMalformed-72]
	_ = x[ErrMalformedPOSTRequest-73]
	_ = x[ErrPOSTFileRequired-74]
	_ = x[ErrSignatureVersionNotSupported-75]
	_ = x[ErrBucketNotEmpty-76]
	_ = x[ErrAllAccessDisabled-77]
	_ = x[ErrPolicyInvalidVersion-78]
	_ = x[ErrMissingFields-79]
	_ = x[ErrMissingCredTag-80]
	_ = x[ErrCredMalformed-81]
	_ = x[ErrInvalidRegion-82]
	_ = x[ErrInvalidServiceS3-83]
	_ = x[ErrInvalidServiceSTS-84]
	_ = x[ErrInvalidRequestVersion-85]
	_ = x[ErrMissingSignTag-86]
	_ = x[ErrMissingSignHeadersTag-87]
	_ = x[ErrMalformedDate-88]
	_ = x[ErrMalformedPresignedDate-89]
	_ = x[ErrMalformedCredentialDate-90]
	_ = x[ErrMalformedExpires-91]
	_ = x[ErrNegativeExpires-92]
	_ = x[ErrAuthHeaderEmpty-93]
	_ = x[ErrExpiredPresignRequest-94]
	_ = x[ErrRequestNotReadyYet-95]
	_ = x[ErrUnsignedHeaders-96]
	_ = x[ErrMissingDateHeader-97]
	_ = x[ErrInvalidQuerySignatureAlgo-98]
	_ = x[ErrInvalidQueryParams-99]
	_ = x[ErrBucketAlreadyOwnedByYou-100]
	_ = x[ErrInvalidDuration-101]
	_ = x[ErrBucketAlreadyExists-102]
	_ = x[ErrMetadataTooLarge-103]
	_ = x[ErrUnsupportedMetadata-104]
	_ = x[ErrMaximumExpires-105]
	_ = x[ErrSlowDown-106]
	_ = x[ErrInvalidPrefixMarker-107]
	_ = x[ErrBadRequest-108]
	_ = x[ErrKeyTooLongError-109]
	_ = x[ErrInvalidBucketObjectLockConfiguration-110]
	_ = x[ErrObjectLockConfigurationNotFound-111]
	_ = x[ErrObjectLockConfigurationNotAllowed-112]
	_ = x[ErrNoSuchObjectLockConfiguration-113]
	_ = x[ErrObjectLocked-114]
	_ = x[ErrInvalidRetentionDate-115]
	_ = x[ErrPastObjectLockRetainDate-116]
	_ = x[ErrUnknownWORMModeDirective-117]
	_ = x[ErrBucketTaggingNotFound-118]
	_ = x[ErrObjectLockInvalidHeaders-119]
	_ = x[ErrInvalidTagDirective-120]
	_ = x[ErrPolicyAlreadyAttached-121]
	_ = x[ErrPolicyNotAttached-122]
	_ = x[ErrInvalidEncryptionMethod-123]
	_ = x[ErrInvalidEncryptionKeyID-124]
	_ = x[ErrInsecureSSECustomerRequest-125]
	_ = x[ErrSSEMultipartEncrypted-126]
	_ = x[ErrSSEEncryptedObject-127]
	_ = x[ErrInvalidEncryptionParameters-128]
	_ = x[ErrInvalidEncryptionParametersSSEC-129]
	_ = x[ErrInvalidSSECustomerAlgorithm-130]
	_ = x[ErrInvalidSSECustomerKey-131]
	_ = x[ErrMissingSSECustomerKey-132]
	_ = x[ErrMissingSSECustomerKeyMD5-133]
	_ = x[ErrSSECustomerKeyMD5Mismatch-134]
	_ = x[ErrInvalidSSECustomerParameters-135]
	_ = x[ErrIncompatibleEncryptionMethod-136]
	_ = x[ErrKMSNotConfigured-137]
	_ = x[ErrKMSKeyNotFoundException-138]
	_ = x[ErrKMSDefaultKeyAlreadyConfigured-139]
	_ = x[ErrNoAccessKey-140]
	_ = x[ErrInvalidToken-141]
	_ = x[ErrEventNotification-142]
	_ = x[ErrARNNotification-143]
	_ = x[ErrRegionNotification-144]
	_ = x[ErrOverlappingFilterNotification-145]
	_ = x[ErrFilterNameInvalid-146]
	_ = x[ErrFilterNamePrefix-147]
	_ = x[ErrFilterNameSuffix-148]
	_ = x[ErrFilterValueInvalid-149]
	_ = x[ErrOverlappingConfigs-150]
	_ = x[ErrUnsupportedNotification-151]
	_ = x[ErrContentSHA256Mismatch-152]
	_ = x[ErrContentChecksumMismatch-153]
	_ = x[ErrStorageFull-154]
	_ = x[ErrRequestBodyParse-155]
	_ = x[ErrObjectExistsAsDirectory-156]
	_ = x[ErrInvalidObjectName-157]
	_ = x[ErrInvalidObjectNamePrefixSlash-158]
	_ = x[ErrInvalidResourceName-159]
	_ = x[ErrServerNotInitialized-160]
	_ = x[ErrOperationTimedOut-161]
	_ = x[ErrClientDisconnected-162]
	_ = x[ErrOperationMaxedOut-163]
	_ = x[ErrInvalidRequest-164]
	_ = x[ErrTransitionStorageClassNotFoundError-165]
	_ = x[ErrInvalidStorageClass-166]
	_ = x[ErrBackendDown-167]
	_ = x[ErrMalformedJSON-168]
	_ = x[ErrAdminNoSuchUser-169]
	_ = x[ErrAdminNoSuchGroup-170]
	_ = x[ErrAdminGroupNotEmpty-171]
	_ = x[ErrAdminNoSuchJob-172]
	_ = x[ErrAdminNoSuchPolicy-173]
	_ = x[ErrAdminPolicyChangeAlreadyApplied-174]
	_ = x[ErrAdminInvalidArgument-175]
	_ = x[ErrAdminInvalidAccessKey-176]
	_ = x[ErrAdminInvalidSecretKey-177]
	_ = x[ErrAdminConfigNoQuorum-178]
	_ = x[ErrAdminConfigTooLarge-179]
	_ = x[ErrAdminConfigBadJSON-180]
	_ = x[ErrAdminNoSuchConfigTarget-181]
	_ = x[ErrAdminConfigEnvOverridden-182]
	_ = x[ErrAdminConfigDuplicateKeys-183]
	_ = x[ErrAdminConfigInvalidIDPType-184]
	_ = x[ErrAdminConfigLDAPValidation-185]
	_ = x[ErrAdminConfigIDPCfgNameAlreadyExists-186]
	_ = x[ErrAdminConfigIDPCfgNameDoesNotExist-187]
	_ = x[ErrAdminCredentialsMismatch-188]
	_ = x[ErrInsecureClientRequest-189]
	_ = x[ErrObjectTampered-190]
	_ = x[ErrSiteReplicationInvalidRequest-191]
	_ = x[ErrSiteReplicationPeerResp-192]
	_ = x[ErrSiteReplicationBackendIssue-193]
	_ = x[ErrSiteReplicationServiceAccountError-194]
	_ = x[ErrSiteReplicationBucketConfigError-195]
	_ = x[ErrSiteReplicationBucketMetaError-196]
	_ = x[ErrSiteReplicationIAMError-197]
	_ = x[ErrSiteReplicationConfigMissing-198]
	_ = x[ErrAdminRebalanceAlreadyStarted-199]
	_ = x[ErrAdminRebalanceNotStarted-200]
	_ = x[ErrAdminBucketQuotaExceeded-201]
	_ = x[ErrAdminNoSuchQuotaConfiguration-202]
	_ = x[ErrHealNotImplemented-203]
	_ = x[ErrHealNoSuchProcess-204]
	_ = x[ErrHealInvalidClientToken-205]
	_ = x[ErrHealMissingBucket-206]
	_ = x[ErrHealAlreadyRunning-207]
	_ = x[ErrHealOverlappingPaths-208]
	_ = x[ErrIncorrectContinuationToken-209]
	_ = x[ErrEmptyRequestBody-210]
	_ = x[ErrUnsupportedFunction-211]
	_ = x[ErrInvalidExpressionType-212]
	_ = x[ErrBusy-213]
	_ = x[ErrUnauthorizedAccess-214]
	_ = x[ErrExpressionTooLong-215]
	_ = x[ErrIllegalSQLFunctionArgument-216]
	_ = x[ErrInvalidKeyPath-217]
	_ = x[ErrInvalidCompressionFormat-218]
	_ = x[ErrInvalidFileHeaderInfo-219]
	_ = x[ErrInvalidJSONType-220]
	_ = x[ErrInvalidQuoteFields-221]
	_ = x[ErrInvalidRequestParameter-222]
	_ = x[ErrInvalidDataType-223]
	_ = x[ErrInvalidTextEncoding-224]
	_ = x[ErrInvalidDataSource-225]
	_ = x[ErrInvalidTableAlias-226]
	_ = x[ErrMissingRequiredParameter-227]
	_ = x[ErrObjectSerializationConflict-228]
	_ = x[ErrUnsupportedSQLOperation-229]
	_ = x[ErrUnsupportedSQLStructure-230]
	_ = x[ErrUnsupportedSyntax-231]
	_ = x[ErrUnsupportedRangeHeader-232]
	_ = x[ErrLexerInvalidChar-233]
	_ = x[ErrLexerInvalidOperator-234]
	_ = x[ErrLexerInvalidLiteral-235]
	_ = x[ErrLexerInvalidIONLiteral-236]
	_ = x[ErrParseExpectedDatePart-237]
	_ = x[ErrParseExpectedKeyword-238]
	_ = x[ErrParseExpectedTokenType-239]
	_ = x[ErrParseExpected2TokenTypes-240]
	_ = x[ErrParseExpectedNumber-241]
	_ = x[ErrParseExpectedRightParenBuiltinFunctionCall-242]
	_ = x[ErrParseExpectedTypeName-243]
	_ = x[ErrParseExpectedWhenClause-244]
	_ = x[ErrParseUnsupportedToken-245]
	_ = x[ErrParseUnsupportedLiteralsGroupBy-246]
	_ = x[ErrParseExpectedMember-247]
	_ = x[ErrParseUnsupportedSelect-248]
	_ = x[ErrParseUnsupportedCase-249]
	_ = x[ErrParseUnsupportedCaseClause-250]
	_ = x[ErrParseUnsupportedAlias-251]
	_ = x[ErrParseUnsupportedSyntax-252]
	_ = x[ErrParseUnknownOperator-253]
	_ = x[ErrParseMissingIdentAfterAt-254]
	_ = x[ErrParseUnexpectedOperator-255]
	_ = x[ErrParseUnexpectedTerm-256]
	_ = x[ErrParseUnexpectedToken-257]
	_ = x[ErrParseUnexpectedKeyword-258]
	_ = x[ErrParseExpectedExpression-259]
	_ = x[ErrParseExpectedLeftParenAfterCast-260]
	_ = x[ErrParseExpectedLeftParenValueConstructor-261]
	_ = x[ErrParseExpectedLeftParenBuiltinFunctionCall-262]
	_ = x[ErrParseExpectedArgumentDelimiter-263]
	_ = x[ErrParseCastArity-264]
	_ = x[ErrParseInvalidTypeParam-265]
	_ = x[ErrParseEmptySelect-266]
	_ = x[ErrParseSelectMissingFrom-267]
	_ = x[ErrParseExpectedIdentForGroupName-268]
	_ = x[ErrParseExpectedIdentForAlias-269]
	_ = x[ErrParseUnsupportedCallWithStar-270]
	_ = x[ErrParseNonUnaryAgregateFunctionCall-271]
	_ = x[ErrParseMalformedJoin-272]
	_ = x[ErrParseExpectedIdentForAt-273]
	_ = x[ErrParseAsteriskIsNotAloneInSelectList-274]
	_ = x[ErrParseCannotMixSqbAndWildcardInSelectList-275]
	_ = x[ErrParseInvalidContextForWildcardInSelectList-276]
	_ = x[ErrIncorrectSQLFunctionArgumentType-277]
	_ = x[ErrValueParseFailure-278]
	_ = x[ErrEvaluatorInvalidArguments-279]
	_ = x[ErrIntegerOverflow-280]
	_ = x[ErrLikeInvalidInputs-281]
	_ = x[ErrCastFailed-282]
	_ = x[ErrInvalidCast-283]
	_ = x[ErrEvaluatorInvalidTimestampFormatPattern-284]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternSymbolForParsing-285]
	_ = x[ErrEvaluatorTimestampFormatPatternDuplicateFields-286]
	_ = x[ErrEvaluatorTimestampFormatPatternHourClockAmPmMismatch-287]
	_ = x[ErrEvaluatorUnterminatedTimestampFormatPatternToken-288]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternToken-289]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternSymbol-290]
	_ = x[ErrEvaluatorBindingDoesNotExist-291]
	_ = x[ErrMissingHeaders-292]
	_ = x[ErrInvalidColumnIndex-293]
	_ = x[ErrAdminConfigNotificationTargetsFailed-294]
	_ = x[ErrAdminProfilerNotEnabled-295]
	_ = x[ErrInvalidDecompressedSize-296]
	_ = x[ErrAddUserInvalidArgument-297]
	_ = x[ErrAdminResourceInvalidArgument-298]
	_ = x[ErrAdminAccountNotEligible-299]
	_ = x[ErrAccountNotEligible-300]
	_ = x[ErrAdminServiceAccountNotFound-301]
	_ = x[ErrPostPolicyConditionInvalidFormat-302]
	_ = x[ErrInvalidChecksum-303]
}

const _APIErrorCode_name = "NoneAccessDeniedBadDigestEntityTooSmallEntityTooLargePolicyTooLargeIncompleteBodyInternalErrorInvalidAccessKeyIDAccessKeyDisabledInvalidBucketNameInvalidDigestInvalidRangeInvalidRangePartNumberInvalidCopyPartRangeInvalidCopyPartRangeSourceInvalidMaxKeysInvalidEncodingMethodInvalidMaxUploadsInvalidMaxPartsInvalidPartNumberMarkerInvalidPartNumberInvalidRequestBodyInvalidCopySourceInvalidMetadataDirectiveInvalidCopyDestInvalidPolicyDocumentInvalidObjectStateMalformedXMLMissingContentLengthMissingContentMD5MissingRequestBodyErrorMissingSecurityHeaderNoSuchBucketNoSuchBucketPolicyNoSuchBucketLifecycleNoSuchLifecycleConfigurationInvalidLifecycleWithObjectLockNoSuchBucketSSEConfigNoSuchCORSConfigurationNoSuchWebsiteConfigurationReplicationConfigurationNotFoundErrorRemoteDestinationNotFoundErrorReplicationDestinationMissingLockRemoteTargetNotFoundErrorReplicationRemoteConnectionErrorReplicationBandwidthLimitErrorBucketRemoteIdenticalToSourceBucketRemoteAlreadyExistsBucketRemoteLabelInUseBucketRemoteArnTypeInvalidBucketRemoteArnInvalidBucketRemoteRemoveDisallowedRemoteTargetNotVersionedErrorReplicationSourceNotVersionedErrorReplicationNeedsVersioningErrorReplicationBucketNeedsVersioningErrorReplicationDenyEditErrorRemoteTargetDenyEditErrorReplicationNoExistingObjectsObjectRestoreAlreadyInProgressNoSuchKeyNoSuchUploadInvalidVersionIDNoSuchVersionNotImplementedPreconditionFailedRequestTimeTooSkewedSignatureDoesNotMatchMethodNotAllowedInvalidPartInvalidPartOrderAuthorizationHeaderMalformedMalformedPOSTRequestPOSTFileRequiredSignatureVersionNotSupportedBucketNotEmptyAllAccessDisabledPolicyInvalidVersionMissingFieldsMissingCredTagCredMalformedInvalidRegionInvalidServiceS3InvalidServiceSTSInvalidRequestVersionMissingSignTagMissingSignHeadersTagMalformedDateMalformedPresignedDateMalformedCredentialDateMalformedExpiresNegativeExpiresAuthHeaderEmptyExpiredPresignRequestRequestNotReadyYetUnsignedHeadersMissingDateHeaderInvalidQuerySignatureAlgoInvalidQueryParamsBucketAlreadyOwnedByYouInvalidDurationBucketAlreadyExistsMetadataTooLargeUnsupportedMetadataMaximumExpiresSlowDownInvalidPrefixMarkerBadRequestKeyTooLongErrorInvalidBucketObjectLockConfigurationObjectLockConfigurationNotFoundObjectLockConfigurationNotAllowedNoSuchObjectLockConfigurationObjectLockedInvalidRetentionDatePastObjectLockRetainDateUnknownWORMModeDirectiveBucketTaggingNotFoundObjectLockInvalidHeadersInvalidTagDirectivePolicyAlreadyAttachedPolicyNotAttachedInvalidEncryptionMethodInvalidEncryptionKeyIDInsecureSSECustomerRequestSSEMultipartEncryptedSSEEncryptedObjectInvalidEncryptionParametersInvalidEncryptionParametersSSECInvalidSSECustomerAlgorithmInvalidSSECustomerKeyMissingSSECustomerKeyMissingSSECustomerKeyMD5SSECustomerKeyMD5MismatchInvalidSSECustomerParametersIncompatibleEncryptionMethodKMSNotConfiguredKMSKeyNotFoundExceptionKMSDefaultKeyAlreadyConfiguredNoAccessKeyInvalidTokenEventNotificationARNNotificationRegionNotificationOverlappingFilterNotificationFilterNameInvalidFilterNamePrefixFilterNameSuffixFilterValueInvalidOverlappingConfigsUnsupportedNotificationContentSHA256MismatchContentChecksumMismatchStorageFullRequestBodyParseObjectExistsAsDirectoryInvalidObjectNameInvalidObjectNamePrefixSlashInvalidResourceNameServerNotInitializedOperationTimedOutClientDisconnectedOperationMaxedOutInvalidRequestTransitionStorageClassNotFoundErrorInvalidStorageClassBackendDownMalformedJSONAdminNoSuchUserAdminNoSuchGroupAdminGroupNotEmptyAdminNoSuchJobAdminNoSuchPolicyAdminPolicyChangeAlreadyAppliedAdminInvalidArgumentAdminInvalidAccessKeyAdminInvalidSecretKeyAdminConfigNoQuorumAdminConfigTooLargeAdminConfigBadJSONAdminNoSuchConfigTargetAdminConfigEnvOverriddenAdminConfigDuplicateKeysAdminConfigInvalidIDPTypeAdminConfigLDAPValidationAdminConfigIDPCfgNameAlreadyExistsAdminConfigIDPCfgNameDoesNotExistAdminCredentialsMismatchInsecureClientRequestObjectTamperedSiteReplicationInvalidRequestSiteReplicationPeerRespSiteReplicationBackendIssueSiteReplicationServiceAccountErrorSiteReplicationBucketConfigErrorSiteReplicationBucketMetaErrorSiteReplicationIAMErrorSiteReplicationConfigMissingAdminRebalanceAlreadyStartedAdminRebalanceNotStartedAdminBucketQuotaExceededAdminNoSuchQuotaConfigurationHealNotImplementedHealNoSuchProcessHealInvalidClientTokenHealMissingBucketHealAlreadyRunningHealOverlappingPathsIncorrectContinuationTokenEmptyRequestBodyUnsupportedFunctionInvalidExpressionTypeBusyUnauthorizedAccessExpressionTooLongIllegalSQLFunctionArgumentInvalidKeyPathInvalidCompressionFormatInvalidFileHeaderInfoInvalidJSONTypeInvalidQuoteFieldsInvalidRequestParameterInvalidDataTypeInvalidTextEncodingInvalidDataSourceInvalidTableAliasMissingRequiredParameterObjectSerializationConflictUnsupportedSQLOperationUnsupportedSQLStructureUnsupportedSyntaxUnsupportedRangeHeaderLexerInvalidCharLexerInvalidOperatorLexerInvalidLiteralLexerInvalidIONLiteralParseExpectedDatePartParseExpectedKeywordParseExpectedTokenTypeParseExpected2TokenTypesParseExpectedNumberParseExpectedRightParenBuiltinFunctionCallParseExpectedTypeNameParseExpectedWhenClauseParseUnsupportedTokenParseUnsupportedLiteralsGroupByParseExpectedMemberParseUnsupportedSelectParseUnsupportedCaseParseUnsupportedCaseClauseParseUnsupportedAliasParseUnsupportedSyntaxParseUnknownOperatorParseMissingIdentAfterAtParseUnexpectedOperatorParseUnexpectedTermParseUnexpectedTokenParseUnexpectedKeywordParseExpectedExpressionParseExpectedLeftParenAfterCastParseExpectedLeftParenValueConstructorParseExpectedLeftParenBuiltinFunctionCallParseExpectedArgumentDelimiterParseCastArityParseInvalidTypeParamParseEmptySelectParseSelectMissingFromParseExpectedIdentForGroupNameParseExpectedIdentForAliasParseUnsupportedCallWithStarParseNonUnaryAgregateFunctionCallParseMalformedJoinParseExpectedIdentForAtParseAsteriskIsNotAloneInSelectListParseCannotMixSqbAndWildcardInSelectListParseInvalidContextForWildcardInSelectListIncorrectSQLFunctionArgumentTypeValueParseFailureEvaluatorInvalidArgumentsIntegerOverflowLikeInvalidInputsCastFailedInvalidCastEvaluatorInvalidTimestampFormatPatternEvaluatorInvalidTimestampFormatPatternSymbolForParsingEvaluatorTimestampFormatPatternDuplicateFieldsEvaluatorTimestampFormatPatternHourClockAmPmMismatchEvaluatorUnterminatedTimestampFormatPatternTokenEvaluatorInvalidTimestampFormatPatternTokenEvaluatorInvalidTimestampFormatPatternSymbolEvaluatorBindingDoesNotExistMissingHeadersInvalidColumnIndexAdminConfigNotificationTargetsFailedAdminProfilerNotEnabledInvalidDecompressedSizeAddUserInvalidArgumentAdminResourceInvalidArgumentAdminAccountNotEligibleAccountNotEligibleAdminServiceAccountNotFoundPostPolicyConditionInvalidFormatInvalidChecksum"

var _APIErrorCode_index = [...]uint16{0, 4, 16, 25, 39, 53, 67, 81, 94, 112, 129, 146, 159, 171, 193, 213, 239, 253, 274, 291, 306, 329, 346, 364, 381, 405, 420, 441, 459, 471, 491, 508, 531, 552, 564, 582, 603, 631, 661, 682, 705, 731, 768, 798, 831, 856, 888, 918, 947, 972, 994, 1020, 1042, 1070, 1099, 1133, 1164, 1201, 1225, 1250, 1278, 1308, 1317, 1329, 1345, 1358, 1372, 1390, 1410, 1431, 1447, 1458, 1474, 1502, 1522, 1538, 1566, 1580, 1597, 1617, 1630, 1644, 1657, 1670, 1686, 1703, 1724, 1738, 1759, 1772, 1794, 1817, 1833, 1848, 1863, 1884, 1902, 1917, 1934, 1959, 1977, 2000, 2015, 2034, 2050, 2069, 2083, 2091, 2110, 2120, 2135, 2171, 2202, 2235, 2264, 2276, 2296, 2320, 2344, 2365, 2389, 2408, 2429, 2446, 2469, 2491, 2517, 2538, 2556, 2583, 2614, 2641, 2662, 2683, 2707, 2732, 2760, 2788, 2804, 2827, 2857, 2868, 2880, 2897, 2912, 2930, 2959, 2976, 2992, 3008, 3026, 3044, 3067, 3088, 3111, 3122, 3138, 3161, 3178, 3206, 3225, 3245, 3262, 3280, 3297, 3311, 3346, 3365, 3376, 3389, 3404, 3420, 3438, 3452, 3469, 3500, 3520, 3541, 3562, 3581, 3600, 3618, 3641, 3665, 3689, 3714, 3739, 3773, 3806, 3830, 3851, 3865, 3894, 3917, 3944, 3978, 4010, 4040, 4063, 4091, 4119, 4143, 4167, 4196, 4214, 4231, 4253, 4270, 4288, 4308, 4334, 4350, 4369, 4390, 4394, 4412, 4429, 4455, 4469, 4493, 4514, 4529, 4547, 4570, 4585, 4604, 4621, 4638, 4662, 4689, 4712, 4735, 4752, 4774, 4790, 4810, 4829, 4851, 4872, 4892, 4914, 4938, 4957, 4999, 5020, 5043, 5064, 5095, 5114, 5136, 5156, 5182, 5203, 5225, 5245, 5269, 5292, 5311, 5331, 5353, 5376, 5407, 5445, 5486, 5516, 5530, 5551, 5567, 5589, 5619, 5645, 5673, 5706, 5724, 5747, 5782, 5822, 5864, 5896, 5913, 5938, 5953, 5970, 5980, 5991, 6029, 6083, 6129, 6181, 6229, 6272, 6316, 6344, 6358, 6376, 6412, 6435, 6458, 6480, 6508, 6531, 6549, 6576, 6608, 6623}

func (i APIErrorCode) String() string {
	if i < 0 || i >= APIErrorCode(len(_APIErrorCode_index)-1) {
		return "APIErrorCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _APIErrorCode_name[_APIErrorCode_index[i]:_APIErrorCode_index[i+1]]
}
