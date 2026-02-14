import { writable, derived } from 'svelte/store';
import { translations } from '$lib/i18n';
import { logger } from '$lib/utils/logger';

const i18nLogger = logger.child('I18n');

export type Language = 'de' | 'en' | 'fr';

function isValidLanguage(lang: string): lang is Language {
	return ['de', 'en', 'fr'].includes(lang);
}

function createI18nStore() {
	const getInitialLanguage = (): Language => {
		if (typeof window !== 'undefined') {
			const saved = localStorage.getItem('savvy_language');
			if (saved && isValidLanguage(saved)) {
				return saved;
			}
		}
		return 'de';
	};

	const { subscribe, set } = writable<Language>(getInitialLanguage());

	return {
		subscribe,
		/**
		 * Set language locally and persist to backend.
		 * Called when the user explicitly changes the language in the UI.
		 */
		setLanguage: (lang: Language) => {
			set(lang);
			if (typeof window !== 'undefined') {
				localStorage.setItem('savvy_language', lang);
			}
			// Sync to backend (fire-and-forget, non-blocking)
			import('$lib/api/profile').then(({ profileApi }) => {
				profileApi.update({ language: lang }).catch((err) => {
					i18nLogger.warn('Failed to sync language to backend:', err);
				});
			});
		},
		/**
		 * Set language from backend without triggering a backend save.
		 * Called when loading user data from /auth/me or /profile.
		 */
		setFromBackend: (lang: string) => {
			if (isValidLanguage(lang)) {
				set(lang);
				if (typeof window !== 'undefined') {
					localStorage.setItem('savvy_language', lang);
				}
			}
		}
	};
}

export const languageStore = createI18nStore();

export const t = derived(languageStore, ($lang) => {
	return (key: string, params?: Record<string, string | number>): string => {
		const keys = key.split('.');
		let value: unknown = translations[$lang];

		for (const k of keys) {
			if (value && typeof value === 'object' && k in value) {
				value = (value as Record<string, unknown>)[k];
			} else {
				i18nLogger.warn(
					`Translation key not found: ${key} for language: ${$lang}`
				);
				return key;
			}
		}

		if (typeof value !== 'string') {
			i18nLogger.warn(`Translation value is not a string: ${key}`);
			return key;
		}

		if (params) {
			return value.replace(/\{(\w+)\}/g, (match, paramKey) => {
				return params[paramKey]?.toString() || match;
			});
		}

		return value;
	};
});

// Map language codes to locale strings for date formatting
export const locale = derived(languageStore, ($lang) => {
	const localeMap: Record<Language, string> = {
		de: 'de-DE',
		en: 'en-US',
		fr: 'fr-FR'
	};
	return localeMap[$lang];
});
