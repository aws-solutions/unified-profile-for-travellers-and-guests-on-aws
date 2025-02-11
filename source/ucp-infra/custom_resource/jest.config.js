module.exports = {
    roots: ['<rootDir>/test'],
    testMatch: ['**/*.spec.ts', '**/*.test.ts'],
    transform: {
        '^.+\\.ts$': 'ts-jest'
    },
    coverageReporters: ['text', ['lcov', { projectRoot: '../' }]],
    setupFiles: []
};
