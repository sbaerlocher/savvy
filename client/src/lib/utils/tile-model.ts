/**
 * TileModel adapter — normalises Card / Voucher / Gift-Card DTOs onto a single
 * shape so one ResourceTile component can render all three types with fixed
 * slot positions. See the ResourceTile design card for the slot inventory.
 *
 * All type-specific logic (amount formatting, computed gift-card status, share
 * label derivation) lives HERE, never in the component.
 */

import type { CardDTO, VoucherDTO, GiftCardDTO } from '$lib/types/api';
import type { BarcodeModalItem } from '$lib/components/dashboard/BarcodeModal.svelte';
import { formatCurrency } from '$lib/utils/currency';
import { t } from '$lib/stores/i18n';
import { get } from 'svelte/store';

/** The unwrapped translate function held by the `t` derived store. */
type Translate = (
	key: string,
	params?: Record<string, string | number>
) => string;

export type TileResourceType = 'card' | 'voucher' | 'gift_card';

/**
 * Compact share state for the tile's icon row. The full-text label
 * ("mit niemandem geteilt" etc.) lives only on the detail page.
 */
export type ShareState =
	| { kind: 'private' }
	| { kind: 'sharedWith'; count: number }
	| { kind: 'sharedFrom'; firstName: string };

export interface TileModel {
	id: string;
	type: TileResourceType;
	href: string;

	merchantName: string;
	merchantColor: string;

	/** Identifier line: Card `program`, Voucher `description`, Gift-Card `notes`. */
	identifier?: string;

	/** Prominent figure: Voucher value / %, Gift-Card balance. Cards have none. */
	amount?: string;

	/** Masked resource number, e.g. `··4821`. */
	maskedNumber?: string;

	/** Expiry badge text, e.g. `7 Tage` / `Abgelaufen`. */
	expiryBadge?: string;
	/** Expiry badge is urgent (few days left) → warm accent instead of muted. */
	expiryUrgent?: boolean;

	/** "ab {date} gültig" — only set when the resource is not yet valid. */
	notYetValid?: string;

	/** Footer-right usage marker (Voucher only): "einmalig" / "mehrfach". */
	usageMarker?: string;

	/** Whether the resource is currently active (drives grayscale overlay). */
	isActive: boolean;
	/** Overlay badge text shown when not active, e.g. "Abgelaufen". */
	statusBadge?: string;

	/** Share slot — ALWAYS present (part of the fixed grid). */
	shareState: ShareState;

	/** Values needed to render / enlarge the barcode. */
	barcodeValue?: string;
	barcodeType: string;
	/** Full payload for the barcode modal (enlarge on tap). */
	barcodeModalItem: BarcodeModalItem;
}

type Owner = { id?: string; first_name?: string; email?: string } | undefined;

/** Days until a date, floored. Negative means already past. */
function daysUntil(iso: string): number {
	const target = new Date(iso.split('T')[0]);
	const today = new Date();
	today.setHours(0, 0, 0, 0);
	return Math.floor((target.getTime() - today.getTime()) / 86_400_000);
}

function localeString(locale: string): string {
	// i18n store holds 'de' | 'en' | 'fr'; toLocaleDateString wants a BCP-47 tag.
	const map: Record<string, string> = {
		de: 'de-CH',
		en: 'en-US',
		fr: 'fr-FR'
	};
	return map[locale.split('-')[0]] ?? 'de-CH';
}

/**
 * Derives the compact share state. Received item (owner ≠ user) → sharedFrom
 * with the owner's first name; own shared-out item → sharedWith count; else
 * private. The tile renders these as an icon row; full text is detail-only.
 */
function shareState(
	owner: Owner,
	sharedWithCount: number,
	currentUserId: string | undefined
): ShareState {
	if (owner && owner.id && owner.id !== currentUserId) {
		return {
			kind: 'sharedFrom',
			firstName: owner.first_name || owner.email || 'User'
		};
	}
	if (sharedWithCount > 0) {
		return { kind: 'sharedWith', count: sharedWithCount };
	}
	return { kind: 'private' };
}

/** Formats an expiry date into a compact badge; flags urgency when ≤ 7 days. */
function expiryBadge(
	tr: Translate,
	iso: string
): { text: string; urgent: boolean } {
	const days = daysUntil(iso);
	if (days < 0) return { text: tr('tile.expired'), urgent: false };
	if (days === 0) return { text: tr('tile.expiresToday'), urgent: true };
	return {
		text: tr('tile.daysRemaining', { count: String(days) }),
		urgent: days <= 7
	};
}

function maskNumber(value: string): string | undefined {
	if (!value) return undefined;
	return `··${value.slice(-4)}`;
}

export function cardToTileModel(
	card: CardDTO,
	currentUserId: string | undefined,
	// Cards have no currency/date to format; param kept for a uniform signature.
	_locale: string
): TileModel {
	const tr = get(t);
	const share = shareState(card.owner, card.shared_with_count, currentUserId);
	const barcodeType = card.barcode_type || 'CODE128';
	return {
		id: card.id,
		type: 'card',
		href: `/cards/${card.id}`,
		merchantName: card.merchant?.name || tr('dashboard.cardType'),
		merchantColor: card.merchant?.color || '#6B7280',
		identifier: card.program || undefined,
		maskedNumber: maskNumber(card.card_number),
		isActive: card.status === 'active',
		statusBadge:
			card.status === 'active' ? undefined : tr(`cards.status.${card.status}`),
		shareState: share,
		barcodeValue: card.card_number || undefined,
		barcodeType,
		barcodeModalItem: {
			type: 'card',
			value: card.card_number,
			barcodeType: card.barcode_type,
			merchantName: card.merchant?.name
		}
	};
}

/** Voucher amount: percentage / fixed / free / points, matching existing display. */
function voucherAmount(
	voucher: VoucherDTO,
	locale: string
): string | undefined {
	const tr = get(t);
	switch (voucher.type) {
		case 'percentage':
			return `${voucher.value}%`;
		case 'fixed_amount':
			return formatCurrency(voucher.value, voucher.currency, locale);
		case 'points_multiplier':
			return `${voucher.value}${tr('vouchers.types.pointsMultiplierDisplay')}`;
		case 'bonus_points':
			return `+${voucher.value}${tr('vouchers.types.bonusPointsDisplay')}`;
		case 'free':
			return tr('vouchers.types.freeDisplay');
		default:
			return undefined;
	}
}

export function voucherToTileModel(
	voucher: VoucherDTO,
	currentUserId: string | undefined,
	locale: string
): TileModel {
	const tr = get(t);
	const share = shareState(
		voucher.owner,
		voucher.shared_with_count,
		currentUserId
	);
	const barcodeType = voucher.barcode_type || 'CODE128';

	// "ab {date} gültig" only while the start date is still in the future.
	let notYetValid: string | undefined;
	if (voucher.valid_from && daysUntil(voucher.valid_from) > 0) {
		notYetValid = tr('tile.notYetValid', {
			date: new Date(voucher.valid_from.split('T')[0]).toLocaleDateString(
				localeString(locale)
			)
		});
	}

	let expiry: { text: string; urgent: boolean } | undefined;
	if (voucher.valid_until) expiry = expiryBadge(tr, voucher.valid_until);

	// single_use → einmalig; every multiple_use_* / one_per_customer → mehrfach.
	let usageMarker: string | undefined;
	if (voucher.usage_limit_type === 'single_use') {
		usageMarker = tr('tile.usageSingle');
	} else if (voucher.usage_limit_type) {
		usageMarker = tr('tile.usageMultiple');
	}

	return {
		id: voucher.id,
		type: 'voucher',
		href: `/vouchers/${voucher.id}`,
		merchantName: voucher.merchant?.name || tr('dashboard.voucherType'),
		merchantColor: voucher.merchant?.color || '#6B7280',
		identifier: voucher.description || undefined,
		amount: voucherAmount(voucher, locale),
		maskedNumber: maskNumber(voucher.code),
		expiryBadge: expiry?.text,
		expiryUrgent: expiry?.urgent,
		notYetValid,
		usageMarker,
		isActive: voucher.status === 'valid',
		statusBadge: voucher.status === 'valid' ? undefined : tr('tile.expired'),
		shareState: share,
		barcodeValue: voucher.code || undefined,
		barcodeType,
		barcodeModalItem: {
			type: 'voucher',
			value: voucher.code,
			barcodeType: voucher.barcode_type,
			merchantName: voucher.merchant?.name,
			displayValue: voucherAmount(voucher, locale),
			description: voucher.description,
			validFrom: voucher.valid_from,
			validUntil: voucher.valid_until,
			status: voucher.status
		}
	};
}

/**
 * Gift-card status is DERIVED, not read from the DTO (balance 0 → depleted,
 * past expiry → expired, else active). Kept in the adapter per the design card.
 */
export function getComputedGiftCardStatus(giftCard: GiftCardDTO): string {
	if (giftCard.current_balance === 0) return 'depleted';
	if (giftCard.expires_at && new Date(giftCard.expires_at) < new Date())
		return 'expired';
	return 'active';
}

export function giftCardToTileModel(
	giftCard: GiftCardDTO,
	currentUserId: string | undefined,
	locale: string
): TileModel {
	const tr = get(t);
	const share = shareState(
		giftCard.owner,
		giftCard.shared_with_count,
		currentUserId
	);
	const barcodeType = giftCard.barcode_type || 'CODE128';
	const status = getComputedGiftCardStatus(giftCard);

	let expiry: { text: string; urgent: boolean } | undefined;
	if (giftCard.expires_at) expiry = expiryBadge(tr, giftCard.expires_at);

	return {
		id: giftCard.id,
		type: 'gift_card',
		href: `/gift-cards/${giftCard.id}`,
		merchantName: giftCard.merchant?.name || tr('dashboard.giftCardType'),
		merchantColor: giftCard.merchant?.color || '#6B7280',
		identifier: giftCard.notes || undefined,
		amount: formatCurrency(giftCard.current_balance, giftCard.currency, locale),
		maskedNumber: maskNumber(giftCard.card_number),
		expiryBadge: expiry?.text,
		expiryUrgent: expiry?.urgent,
		isActive: status === 'active',
		statusBadge:
			status === 'active' ? undefined : tr(`tile.status.${status}` as string),
		shareState: share,
		barcodeValue: giftCard.card_number || undefined,
		barcodeType,
		barcodeModalItem: {
			type: 'gift_card',
			value: giftCard.card_number,
			barcodeType: giftCard.barcode_type,
			merchantName: giftCard.merchant?.name,
			pin: giftCard.pin,
			balance: giftCard.current_balance.toFixed(2),
			currency: giftCard.currency,
			expiresAt: giftCard.expires_at || undefined,
			status
		}
	};
}
