// @ts-check
// Protractor configuration file, see link for more information
// https://github.com/angular/protractor/blob/master/lib/config.ts

const { SpecReporter } = require('jasmine-spec-reporter');

/**
 * @type { import("protractor").Config }
 */
exports.config = {
  allScriptsTimeout: 50000,
  getPageTimeout: 10000,
  specs: ['./src/features/**/*.feature'],
  capabilities: {
    browserName: 'chrome',
    chromeOptions: {
      args: ["--headless", "--no-sandbox", "--disable-dev-shm-usage", "--window-size=1843,984"]
    }
  },
  directConnect: true,
  baseUrl: 'http://localhost:4200/',
  framework: 'custom',
  frameworkPath: require.resolve('protractor-cucumber-framework'),
  cucumberOpts: {
    require: ['./src/steps/**/*.steps.ts'],
  },
  onPrepare() {
    require('ts-node').register({
      project: require('path').join(__dirname, './tsconfig.json')
    });
  }
};