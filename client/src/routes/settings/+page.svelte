<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';

	// Legacy route: the account destination is /profile on every platform
	// (Android renders the M3 settings screen there). Old tab URLs still map
	// to their dedicated pages.
	onMount(() => {
		const urlTab = page.url.searchParams.get('tab');
		if (urlTab === 'security') {
			goto(resolve('security'), { replace: true });
			return;
		}
		if (urlTab === 'notifications') {
			goto(resolve('notifications/settings'), { replace: true });
			return;
		}
		goto(resolve('profile'), { replace: true });
	});
</script>
