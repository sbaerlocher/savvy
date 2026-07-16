<script lang="ts">
	/* global __APP_VERSION__ -- compile-time constant injected by Vite `define` (see vite.config.ts, typed in src/app.d.ts) */
	import { resolve } from '$app/paths';
	import { t } from '$lib/stores/i18n';
</script>

<footer
	class="bg-gradient-to-b from-surface-1 to-border-soft border-t border-border mt-12 hidden sm:block"
>
	<div class="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
		<div class="grid grid-cols-1 md:grid-cols-4 gap-8">
			<!-- App Info -->
			<div class="space-y-4">
				<h3 class="text-lg font-bold text-text">{$t('common.appName')}</h3>
				<p class="text-sm text-text-muted leading-relaxed">
					{$t('footer.description')}
				</p>
				<div class="flex space-x-3">
					<a
						href="https://github.com/sbaerlocher/savvy"
						target="_blank"
						rel="noopener noreferrer"
						class="text-text-subtle hover:text-text-ink2 transition-colors"
						aria-label={$t('aria.githubLink')}
					>
						<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
							<path
								fill-rule="evenodd"
								d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
								clip-rule="evenodd"
							></path>
						</svg>
					</a>
				</div>
			</div>

			<!-- Quick Links -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-text uppercase tracking-wider">
					{$t('footer.quickLinks')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href={resolve('/dashboard')}
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('nav.start')}
						</a>
					</li>
					<li>
						<a
							href={resolve('/wallet')}
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('nav.wallet')}
						</a>
					</li>
					<li>
						<a
							href={resolve('/profile')}
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('nav.profile')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Resources -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-text uppercase tracking-wider">
					{$t('footer.resources')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy#readme"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('footer.documentation')}
						</a>
					</li>
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy/issues"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('footer.reportBug')}
						</a>
					</li>
					<li>
						<a
							href="https://github.com/sbaerlocher/savvy/releases"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							{$t('footer.changelog')}
						</a>
					</li>
				</ul>
			</div>

			<!-- Tech Stack -->
			<div class="space-y-4">
				<h3 class="text-sm font-semibold text-text uppercase tracking-wider">
					{$t('footer.builtWith')}
				</h3>
				<ul class="space-y-2">
					<li>
						<a
							href="https://svelte.dev"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							SvelteKit
						</a>
					</li>
					<li>
						<a
							href="https://echo.labstack.com"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							Echo Framework
						</a>
					</li>
					<li>
						<a
							href="https://tailwindcss.com"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							Tailwind CSS
						</a>
					</li>
					<li>
						<a
							href="https://gorm.io"
							target="_blank"
							rel="noopener noreferrer"
							class="text-sm text-text-muted hover:text-accent transition-colors"
						>
							GORM
						</a>
					</li>
				</ul>
			</div>
		</div>

		<!-- Bottom Bar -->
		<div class="mt-8 pt-8 border-t border-border-field">
			<div
				class="flex flex-col md:flex-row justify-between items-center space-y-4 md:space-y-0"
			>
				<div
					class="flex flex-col md:flex-row items-center gap-3 text-sm text-text-subtle"
				>
					<p>{$t('footer.copyright')}</p>
					<span class="hidden md:inline">•</span>
					<p class="flex items-center gap-1.5 font-mono">
						<svg
							class="w-4 h-4"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
							></path>
						</svg>
						<span>v{__APP_VERSION__}</span>
					</p>
				</div>
				<div class="flex items-center space-x-2 text-sm text-text-subtle">
					<span>{$t('footer.developedWith')}</span>
					<svg
						class="w-4 h-4 text-danger-500"
						fill="currentColor"
						viewBox="0 0 20 20"
					>
						<path
							fill-rule="evenodd"
							d="M3.172 5.172a4 4 0 015.656 0L10 6.343l1.172-1.171a4 4 0 115.656 5.656L10 17.657l-6.828-6.829a4 4 0 010-5.656z"
							clip-rule="evenodd"
						></path>
					</svg>
					<span>{$t('footer.inSwitzerland')}</span>
				</div>
			</div>
		</div>
	</div>
</footer>
