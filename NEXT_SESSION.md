# Next Session: Status Localization Feature

## Context
We have just completed enhancing the competitor status mapping with dynamic statuses in PR #62. The status codes are now properly mapped to human-readable strings, and we have "Waiting to Start" and "Running" statuses that are determined dynamically based on competitor state.

## New Feature Request
Implement a localization feature for status language display, starting with support for English (default) and Danish. This should be configured through a CLI flag when starting the application.

## Requirements

### 1. CLI Flag
- Add a new CLI flag `--language` or `-l` to specify the language
- Default to English if not specified
- Support values: `en` (English), `da` (Danish)
- Example: `./meos-graphics --language da`

### 2. Status Translations
The following statuses need to be translated:

| Status Code | English | Danish |
|-------------|---------|---------|
| 0 | Unknown | Ukendt |
| 1 | Approved | Godkendt |
| 3 | Miss Punch | Fejlstempel |
| 4 | Not Finished | Ikke Gennemført |
| 5 | Disqualified | Diskvalificeret |
| 6 | Max. Time | Max. Tid |
| 20 | Not Started | Ikke Startet |
| 21 | Cancelled | Annulleret |
| 99 | Not Competing | Deltager Ikke |
| 1000 | Waiting to Start | Venter på Start |
| 1001 | Running | Løber |

Additional status strings used in the service layer:
- "OK" → "OK" (same in Danish)
- "DNF" → "DNF" (same in Danish)
- "MP" → "FS" (Fejlstempel)
- "DSQ" → "DSQ" (same in Danish)
- "OT" → "MT" (Max Tid)
- "DNS" → "DNS" (same in Danish)

### 3. Implementation Approach
1. Create a localization package (`internal/i18n`) to handle translations
2. Modify the service layer to use localized strings based on the configured language
3. Update the `GetStatusDescription()` function to support multiple languages
4. Ensure thread-safe access to the language configuration
5. Consider future extensibility for additional languages

### 4. Testing
- Unit tests for the localization package
- Integration tests to verify correct status strings are returned based on language setting
- Manual testing with both English and Danish configurations

### 5. Documentation
- Update CLI documentation to include the new `--language` flag
- Add examples showing how to run the application in different languages
- Document how to add support for additional languages in the future

## Technical Considerations
- The language setting should be application-wide (not per-request)
- Consider using a proper i18n library or implement a simple translation map
- Ensure the solution is extensible for future language additions
- The API responses should return localized status strings based on the server's language configuration

## Files to Modify/Create
- `cmd/meos-graphics/main.go` - Add CLI flag
- `internal/i18n/` - New package for localization
- `internal/service/service.go` - Update to use localized strings
- `internal/meos/adapter.go` - Update GetStatusDescription to support languages
- Tests for the new functionality
- Documentation updates

## Note
Remember that PR #62 introduced the status mapping enhancements and is currently under review. This localization feature should be implemented after that PR is merged to avoid conflicts.