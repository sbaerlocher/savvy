import type { FullConfig } from '@playwright/test';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

// Configuration constants
const CONFIG = {
	POSTGRES_TIMEOUT_MS: parseInt(
		process.env.E2E_POSTGRES_TIMEOUT_MS || '30000',
		10
	), // 30s
	APP_TIMEOUT_MS: parseInt(process.env.E2E_APP_TIMEOUT_MS || '90000', 10), // 90s
	POLL_INTERVAL_MS: parseInt(process.env.E2E_POLL_INTERVAL_MS || '2000', 10), // 2s
	APP_STARTUP_DELAY_MS: parseInt(
		process.env.E2E_APP_STARTUP_DELAY_MS || '4000',
		10
	), // 4s
	EXEC_TIMEOUT_MS: parseInt(process.env.E2E_EXEC_TIMEOUT_MS || '300000', 10) // 5min for docker build
} as const;

/**
 * Formats elapsed time for logging
 */
function formatElapsed(startTime: number): string {
	const elapsed = Date.now() - startTime;
	return `${(elapsed / 1000).toFixed(1)}s`;
}

/**
 * Validates that Docker is available and running
 */
async function validateDockerAvailable(): Promise<void> {
	try {
		await execAsync('docker info', { timeout: 5000 });
	} catch (error) {
		throw new Error(
			'Docker is not available or not running. Please start Docker Desktop and try again.\n' +
				'Error: ' +
				(error instanceof Error ? error.message : String(error))
		);
	}
}

/**
 * Executes a Docker Compose command with timeout and error handling
 */
async function executeDockerCommand(
	command: string,
	description: string,
	timeoutMs: number = CONFIG.EXEC_TIMEOUT_MS
): Promise<string> {
	try {
		const { stdout, stderr } = await execAsync(command, { timeout: timeoutMs });
		if (stderr && !stderr.includes('Pulling') && !stderr.includes('Building')) {
			console.warn(`   ⚠️  Warning during ${description}:`, stderr);
		}
		return stdout;
	} catch (error) {
		const errorMessage = error instanceof Error ? error.message : String(error);
		throw new Error(`Failed to ${description}: ${errorMessage}`);
	}
}

/**
 * Waits for a service to become healthy with configurable timeout
 */
async function waitForServiceHealth(
	serviceName: string,
	timeoutMs: number,
	stepNumber: string
): Promise<void> {
	const startTime = Date.now();
	const maxAttempts = Math.ceil(timeoutMs / CONFIG.POLL_INTERVAL_MS);

	console.log(
		`   ${stepNumber}  Waiting for ${serviceName} to be ready (timeout: ${timeoutMs / 1000}s)...`
	);

	for (let attempt = 0; attempt < maxAttempts; attempt++) {
		await new Promise((resolve) =>
			setTimeout(resolve, CONFIG.POLL_INTERVAL_MS)
		);

		try {
			const stdout = await executeDockerCommand(
				`docker compose --profile e2e ps ${serviceName}`,
				`check ${serviceName} status`,
				10000
			);

			if (stdout.includes('healthy')) {
				console.log(
					`   ✓ ${serviceName} is healthy (took ${formatElapsed(startTime)})`
				);
				return;
			}

			// Log progress every 10 seconds
			if (attempt > 0 && attempt % 5 === 0) {
				console.log(
					`   ⏳ Still waiting for ${serviceName}... (${formatElapsed(startTime)})`
				);
			}
		} catch (error) {
			// Continue polling on transient errors
			if (attempt === maxAttempts - 1) {
				throw error;
			}
		}
	}

	// Timeout reached - capture logs for debugging
	try {
		const logs = await executeDockerCommand(
			`docker compose --profile e2e logs --tail=50 ${serviceName}`,
			`get ${serviceName} logs`,
			10000
		);
		console.error(`\n❌ ${serviceName} logs (last 50 lines):\n${logs}\n`);
	} catch (logError) {
		console.error(`   ⚠️  Could not retrieve ${serviceName} logs:`, logError);
	}

	throw new Error(
		`Timeout waiting for ${serviceName} to become healthy after ${timeoutMs / 1000}s. ` +
			`Check the logs above for details.`
	);
}

/**
 * Cleans up Docker environment on failure
 */
async function cleanupOnFailure(): Promise<void> {
	try {
		console.log('\n🧹 Cleaning up Docker environment...');
		await execAsync('docker compose --profile e2e down -v', { timeout: 30000 });
		console.log('   ✓ Cleanup completed');
	} catch (error) {
		console.error('   ⚠️  Cleanup failed:', error);
	}
}

/**
 * Checks if setup should be skipped based on command-line arguments
 */
function shouldSkipSetup(config: FullConfig): boolean {
	// Check for explicit argument flags (not substrings in paths)
	// Use process.argv array to avoid false positives like "savvy" containing "-v"
	const isListMode = process.argv.includes('--list');
	const isHelpMode =
		process.argv.includes('--help') || process.argv.includes('-h');
	const isVersionMode =
		process.argv.includes('--version') || process.argv.includes('-v');
	const skipSetup = process.env.SKIP_E2E_SETUP === 'true';

	return isListMode || isHelpMode || isVersionMode || skipSetup;
}

/**
 * Global setup - runs once before all tests
 * Ensures clean state: stops containers, removes volumes, starts fresh environment
 */
async function globalSetup(config: FullConfig): Promise<void> {
	const globalStartTime = Date.now();

	// Skip setup for info-only commands
	if (shouldSkipSetup(config)) {
		console.log(
			'ℹ️  Skipping global setup (info-only command or SKIP_E2E_SETUP=true)'
		);
		return;
	}

	console.log('🧹 Global Setup: Preparing E2E test environment\n');
	console.log('Configuration:');
	console.log(`   - PostgreSQL timeout: ${CONFIG.POSTGRES_TIMEOUT_MS / 1000}s`);
	console.log(`   - App timeout: ${CONFIG.APP_TIMEOUT_MS / 1000}s`);
	console.log(`   - Poll interval: ${CONFIG.POLL_INTERVAL_MS / 1000}s\n`);

	try {
		// Step 0: Validate Docker is available
		console.log('   0️⃣  Validating Docker is available...');
		await validateDockerAvailable();
		console.log('   ✓ Docker is running\n');

		// Step 1: Stop all containers and remove volumes
		console.log('   1️⃣  Stopping containers and removing volumes...');
		await executeDockerCommand(
			'docker compose down -v',
			'stop containers and remove volumes'
		);
		console.log('   ✓ Cleanup completed\n');

		// Step 2: Start PostgreSQL
		console.log('   2️⃣  Starting PostgreSQL...');
		await executeDockerCommand(
			'docker compose --profile e2e up -d postgres',
			'start PostgreSQL'
		);
		console.log('   ✓ PostgreSQL container started\n');

		// Step 3: Wait for PostgreSQL to be healthy
		await waitForServiceHealth('postgres', CONFIG.POSTGRES_TIMEOUT_MS, '3️⃣');
		console.log('');

		// Step 4: Build app-e2e with cache
		console.log('   4️⃣  Building app-e2e (this may take a few minutes)...');
		const buildStartTime = Date.now();
		await executeDockerCommand(
			'docker compose --profile e2e build app-e2e',
			'build app-e2e',
			CONFIG.EXEC_TIMEOUT_MS
		);
		console.log(
			`   ✓ Build completed (took ${formatElapsed(buildStartTime)})\n`
		);

		// Step 5: Start app-e2e (triggers migration + seed)
		console.log('   5️⃣  Starting app-e2e (production build with auto-seed)...');
		await executeDockerCommand(
			'docker compose --profile e2e up -d app-e2e',
			'start app-e2e'
		);
		console.log(`   ✓ app-e2e container started`);
		console.log(
			`   ⏳ Waiting ${CONFIG.APP_STARTUP_DELAY_MS / 1000}s for migrations and seeding...\n`
		);
		await new Promise((resolve) =>
			setTimeout(resolve, CONFIG.APP_STARTUP_DELAY_MS)
		);

		// Step 6: Wait for app-e2e to be healthy
		await waitForServiceHealth('app-e2e', CONFIG.APP_TIMEOUT_MS, '6️⃣');

		console.log('\n✅ E2E environment ready');
		console.log(`   Total setup time: ${formatElapsed(globalStartTime)}\n`);
	} catch (error) {
		console.error(
			'\n❌ Global setup failed:',
			error instanceof Error ? error.message : error
		);

		// Cleanup on failure
		await cleanupOnFailure();

		// Provide helpful troubleshooting tips
		console.error('\n💡 Troubleshooting tips:');
		console.error(
			'   1. Check Docker Desktop is running and has enough resources'
		);
		console.error('   2. Run: docker compose --profile e2e logs app-e2e');
		console.error('   3. Run: docker compose --profile e2e logs postgres');
		console.error('   4. Try manual cleanup: docker compose down -v');
		console.error('   5. Check for port conflicts (5432, 8080)');
		console.error('');

		throw error;
	}
}

export default globalSetup;
