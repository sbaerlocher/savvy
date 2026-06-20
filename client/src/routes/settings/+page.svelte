<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	onMount(() => {
		// Legacy redirect: old tab URLs → /profile
		const urlTab = $page.url.searchParams.get('tab');
		if (urlTab === 'profile') {
			goto(resolve('/profile'), { replaceState: true });
			return;
		}
		if (urlTab === 'security') {
			goto(resolve('/security'), { replaceState: true });
			return;
		}
		goto(resolve('/notifications'), { replaceState: true });
	});
</script>
