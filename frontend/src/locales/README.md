# Locale Key Conventions

Use domain-prefixed translation keys with the format:

- domain.section.label

Examples:

- nav.dashboard
- layout.sidebarSubtitle
- users.form.submit
- errors.validation.required

Rules:

- Keep keys stable; update values, not key names, when copy changes.
- Do not hardcode user-facing strings in Vue components.
- Add keys to both `vi.json` and `en.json` in the same change.
- Use lowercase camelCase for terminal segments.
- Group by feature/domain first (nav, users, matches, settlements, fund, config, errors).

Validation:

- A translation consistency check should fail CI if keys are missing in either locale.
- Local hardcode audit command for review:
	`grep -RInE 'title="|label="|empty-text="|active-text="|inactive-text="|confirmButtonText:|cancelButtonText:|>[^<{]*[A-Za-z][^<{]*<' frontend/src --include='*.vue' --include='*.ts'`

Fallback and missing-key policy:

- `vi` is the default and fallback locale.
- Missing keys should log warnings in development.
- If a key is still missing after fallback, UI displays the key string so missing entries are visible.
