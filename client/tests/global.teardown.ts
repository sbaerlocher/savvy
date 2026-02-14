import type { FullConfig } from '@playwright/test';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

// Configuration constants
const CONFIG = {
	CLEANUP_TIMEOUT_MS: parseInt(
		process.env.E2E_CLEANUP_TIMEOUT_MS || '30000',
		10
	), // 30s
	KEEP_CONTAINERS: process.env.E2E_KEEP_CONTAINERS === 'true', // Keep containers for debugging
	REMOVE_VOLUMES: process.env.E2E_REMOVE_VOLUMES !== 'false' // Remove volumes by default
} as const;

/**
 * Formats elapsed time for logging
 */
function formatElapsed(startTime: number): string {
	const elapsed = Date.now() - startTime;
	return `${(elapsed / 1000).toFixed(1)}s`;
}

/**
 * Checks if teardown should be skipped based on command-line arguments
 */
function shouldSkipTeardown(config: FullConfig): boolean {
	const args = process.argv.join(' ');
	const isListMode = args.includes('--list') || config.projects.length === 0;
	const isHelpMode = args.includes('--help') || args.includes('-h');
	const isVersionMode = args.includes('--version') || args.includes('-v');
	const skipTeardown = process.env.SKIP_E2E_SETUP === 'true';

	return isListMode || isHelpMode || isVersionMode || skipTeardown;
}

/**
 * Executes a Docker Compose command with timeout and error handling
 */
async function executeDockerCommand(
	command: string,
	description: string,
	timeoutMs: number = CONFIG.CLEANUP_TIMEOUT_MS
): Promise<string> {
	try {
		const { stdout, stderr } = await execAsync(command, { timeout: timeoutMs });
		if (
			stderr &&
			!stderr.includes('Stopping') &&
			!stderr.includes('Removing')
		) {
			console.warn(`   ⚠️  Warning during ${description}:`, stderr);
		}
		return stdout;
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : String(error);
		throw new Error(`Failed to ${description}: ${errorMessage}`);
	}
}

/**
 * Displays container logs for debugging
 */
async function displayContainerLogs(): Promise<void> {
	try {
		console.log('\n📋 Container logs (last 20 lines each):\n');

		// PostgreSQL logs
		try {
			const postgresLogs = await executeDockerCommand(
				'docker compose --profile e2e logs --tail=20 postgres',
				'get PostgreSQL logs',
				10000
			);
			console.log('PostgreSQL:');
			console.log(postgresLogs);
		} catch (error) {
			console.log('PostgreSQL: (no logs available)');
		}

		// App logs
		try {
			const appLogs = await executeDockerCommand(
				'docker compose --profile e2e logs --tail=20 app-e2e',
				'get app-e2e logs',
				10000
			);
			console.log('\napp-e2e:');
			console.log(appLogs);
		} catch (error) {
			console.log('\napp-e2e: (no logs available)');
		}
	} catch (error) {
		console.error('   ⚠️  Could not display container logs:', error);
	}
}

/**
 * Stops and removes containers based on configuration
 */
async function cleanupContainers(): Promise<void> {
	const volumeFlag = CONFIG.REMOVE_VOLUMES ? '-v' : '';
	const command = `docker compose --profile e2e down ${volumeFlag}`.trim();

	console.log(
		`   🗑️  Stopping containers${CONFIG.REMOVE_VOLUMES ? ' and removing volumes' : ''}...`
	);
	await executeDockerCommand(command, 'stop and remove containers');
	console.log('   ✓ Cleanup completed');
}

/**
 * Global teardown - runs once after all tests
 * Cleans up test environment and optionally displays logs
 */
async function globalTeardown(config: FullConfig): Promise<void> {
	const teardownStartTime = Date.now();

	// Skip teardown for info-only commands
	if (shouldSkipTeardown(config)) {
		console.log(
			'ℹ️  Skipping global teardown (info-only command or SKIP_E2E_SETUP=true)'
		);
		return;
	}

	console.log('\n🧹 Global Teardown: Cleaning up E2E test environment\n');

	// Show configuration
	if (CONFIG.KEEP_CONTAINERS) {
		console.log('⚠️  E2E_KEEP_CONTAINERS=true: Containers will NOT be removed');
		console.log('   Run manually: docker compose --profile e2e down -v\n');
		return;
	}

	try {
		// Display logs if verbose mode is enabled
		if (process.env.E2E_VERBOSE_LOGS === 'true') {
			await displayContainerLogs();
		}

		// Cleanup containers and volumes
		await cleanupContainers();

		console.log(
			`\n✅ Teardown completed (took ${formatElapsed(teardownStartTime)})\n`
		);
	} catch (error) {
		console.error(
			'\n❌ Global teardown failed:',
			error instanceof Error ? error.message : error
		);

		// Provide helpful troubleshooting tips
		console.error('\n💡 Troubleshooting tips:');
		console.error(
			'   1. Try manual cleanup: docker compose --profile e2e down -v'
		);
		console.error(
			'   2. Check for running containers: docker compose --profile e2e ps'
		);
		console.error(
			'   3. Force remove: docker compose --profile e2e rm -f -s -v'
		);
		console.error(
			'   4. Remove orphan containers: docker compose down --remove-orphans'
		);
		console.error('');

		// Don't throw - allow tests to complete even if cleanup fails
		console.warn('⚠️  Continuing despite teardown failure...\n');
	}
}

export default globalTeardown;
