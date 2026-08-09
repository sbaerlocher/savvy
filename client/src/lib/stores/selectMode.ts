import { writable } from 'svelte/store';

/**
 * True while a list screen is in batch-selection mode. Both native platforms
 * rearrange their navigation chrome in that state, so the layout has to know:
 * Android swaps in the M3 contextual top app bar and drops nav bar plus FAB,
 * iOS hides the bottom nav because the floating batch bar takes its slot. A
 * store rather than a prop, since WalletView sits several levels below.
 */
export const selectModeActive = writable(false);
