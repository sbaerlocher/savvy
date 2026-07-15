// Svelte action: relocate the node to document.body (escape overflow/stacking
// contexts) and clean it up on destroy.
export function portal(node: HTMLElement) {
	node.style.margin = '0';
	document.body.appendChild(node);

	return {
		destroy() {
			if (node.parentNode) {
				node.parentNode.removeChild(node);
			}
		}
	};
}
