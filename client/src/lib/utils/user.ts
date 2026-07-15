/**
 * A minimal user shape carrying the fields relevant to display naming.
 * Every field is optional so callers can hand in partial API objects.
 */
export interface DisplayUser {
	first_name?: string | null;
	last_name?: string | null;
	email?: string | null;
}

/**
 * Formats a user's display name with the precedence
 * `"first last"` → `first` → `email` → `fallback`.
 *
 * The i18n "unknown user" case stays out of this helper: callers resolve the
 * translated string themselves and pass it as `fallback`, keeping this pure.
 *
 * Note: this always concatenates `last_name` when both names are present, so it
 * is NOT a drop-in for sites that intentionally show first-name-only, nor for
 * sites that fall back to a bare last name when only `last_name` exists.
 */
export function formatUserName(
	user: DisplayUser | null | undefined,
	fallback = ''
): string {
	if (user?.first_name && user?.last_name) {
		return `${user.first_name} ${user.last_name}`;
	}
	if (user?.first_name) return user.first_name;
	return user?.email || fallback;
}
