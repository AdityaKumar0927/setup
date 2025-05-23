/** @type {import('eslint').Linter.Config} */

module.exports = {
  root: true,
  extends: ["@remix-run/eslint-config", "@remix-run/eslint-config/node"],
  rules: {
    // Example custom ESLint rules:
    "no-console": "warn",
    "@typescript-eslint/no-unused-vars": "warn"
  }
};
