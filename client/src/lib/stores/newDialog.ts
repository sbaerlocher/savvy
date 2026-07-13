import { writable } from 'svelte/store';

/**
 * Controls the global "New" (TypeChoiceDialog) visibility. Lives in a store so
 * both the layout and any page header (MobileHeaderActions) can open it without
 * threading callbacks through the component tree.
 */
export const showNewDialog = writable(false);
