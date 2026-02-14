import { writable } from 'svelte/store';
import { logger } from '$lib/utils/logger';

export interface AppConfig {
	oauth: {
		enabled: boolean;
		login_url?: string;
	};
	features: {
		cards: boolean;
		vouchers: boolean;
		gift_cards: boolean;
	};
	local_login_enabled: boolean;
	registration_enabled: boolean;
	smtp_enabled: boolean;
	push_enabled: boolean;
	two_factor_enabled: boolean;
	is_development: boolean;
}

const DEFAULT_CONFIG: AppConfig = {
	oauth: { enabled: false },
	features: { cards: true, vouchers: true, gift_cards: true },
	local_login_enabled: true,
	registration_enabled: true,
	smtp_enabled: false,
	push_enabled: false,
	two_factor_enabled: false,
	is_development: false
};

function createConfigStore() {
	const { subscribe, set, update } = writable<AppConfig>(DEFAULT_CONFIG);
	let resolveLoaded: () => void;
	const loaded = new Promise<void>((resolve) => {
		resolveLoaded = resolve;
	});

	return {
		subscribe,
		loaded,
		async load(): Promise<void> {
			try {
				const response = await fetch('/api/v1/config');
				if (!response.ok) {
					logger.error('Failed to load config:', response.statusText);
					return;
				}

				const config: AppConfig = await response.json();

				logger.info('Config loaded from backend', config);

				set(config);
			} catch (error) {
				logger.error('Failed to load config:', error);
			} finally {
				resolveLoaded();
			}
		},
		set,
		update
	};
}

export const configStore = createConfigStore();
