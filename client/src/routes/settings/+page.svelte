<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';

	onMount(() => {
		// Legacy redirect: old tab URLs → /profile
		const urlTab = $page.url.searchParams.get('tab');
		if (urlTab === 'profile') {
			goto('/profile', { replaceState: true });
			return;
		}
		if (urlTab === 'security') {
			goto('/security', { replaceState: true });
			return;
		}
		goto('/notifications', { replaceState: true });
	});
</script>
