/**
 * Client-side logger with configurable log levels
 * Mirrors backend log levels: DEBUG, INFO, WARN, ERROR
 */

type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

// Log level priority (higher = more severe)
const LOG_LEVEL_PRIORITY: Record<LogLevel, number> = {
	DEBUG: 0,
	INFO: 1,
	WARN: 2,
	ERROR: 3
};

class Logger {
	private currentLevel: LogLevel = import.meta.env.DEV ? 'DEBUG' : 'WARN';
	private prefix: string = '';

	/**
	 * Create a child logger with a prefix
	 * Useful for module-specific logging (e.g., "API", "Store", "Component")
	 */
	child(prefix: string): Logger {
		const child = new Logger();
		child.currentLevel = this.currentLevel;
		child.prefix = prefix;
		return child;
	}

	/**
	 * Check if a log level should be logged
	 */
	private shouldLog(level: LogLevel): boolean {
		return LOG_LEVEL_PRIORITY[level] >= LOG_LEVEL_PRIORITY[this.currentLevel];
	}

	/**
	 * Format log message with prefix and emoji
	 */
	private format(level: LogLevel, message: string): string {
		const emoji = {
			DEBUG: '🔍',
			INFO: 'ℹ️',
			WARN: '⚠️',
			ERROR: '❌'
		}[level];

		const prefix = this.prefix ? `[${this.prefix}]` : '';
		return `${emoji} ${prefix} ${message}`;
	}

	/**
	 * Log a debug message
	 */
	debug(message: string, ...args: unknown[]): void {
		if (this.shouldLog('DEBUG')) {
			console.debug(this.format('DEBUG', message), ...args);
		}
	}

	/**
	 * Log an info message
	 */
	info(message: string, ...args: unknown[]): void {
		if (this.shouldLog('INFO')) {
			console.info(this.format('INFO', message), ...args);
		}
	}

	/**
	 * Log a warning message
	 */
	warn(message: string, ...args: unknown[]): void {
		if (this.shouldLog('WARN')) {
			console.warn(this.format('WARN', message), ...args);
		}
	}

	/**
	 * Log an error message
	 */
	error(message: string, ...args: unknown[]): void {
		if (this.shouldLog('ERROR')) {
			console.error(this.format('ERROR', message), ...args);
		}
	}

	/**
	 * Log API requests (special case)
	 */
	api(method: string, endpoint: string, status?: number): void {
		if (this.shouldLog('DEBUG')) {
			const statusText = status ? ` → ${status}` : '';
			console.debug(
				this.format('DEBUG', `API ${method} ${endpoint}${statusText}`)
			);
		}
	}
}

// Global logger instance
export const logger = new Logger();

// Default export for convenience
export default logger;
