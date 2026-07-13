<script lang="ts">
	import { authStore } from '$lib/stores/auth';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { onMount } from 'svelte';

	onMount(() => {
		// Redirect to dashboard if logged in, otherwise to login
		if ($authStore.isAuthenticated) {
			goto(resolve('/dashboard'));
		} else {
			goto(resolve('/login'));
		}
	});
</script>

<div class="min-h-screen flex items-center justify-center">
	<div class="text-center">
		<h1 class="text-4xl font-bold text-accent mb-4">Savvy</h1>
		<p class="text-text-muted">Lade...</p>
	</div>
</div>
