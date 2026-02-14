import { api } from './client';

export const pushApi = {
	async getVAPIDKey(): Promise<{ public_key: string }> {
		return api.get('/push/vapid-key');
	},

	async subscribe(subscription: PushSubscription): Promise<void> {
		const json = subscription.toJSON();

		if (!json.keys?.p256dh || !json.keys?.auth) {
			throw new Error('Push subscription missing keys');
		}

		await api.post('/push/subscribe', {
			endpoint: subscription.endpoint,
			keys: {
				p256dh: json.keys.p256dh,
				auth: json.keys.auth
			}
		});
	},

	async unsubscribe(endpoint: string): Promise<void> {
		await api.post('/push/unsubscribe', { endpoint });
	}
};
