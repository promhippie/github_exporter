// yaml-language-server: $schema=https://json.schemastore.org/commitlintrc.json
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [
      2,
      'always',
      [
        'feat',
        'fix',
        'docs',
        'style',
        'refactor',
        'perf',
        'test',
        'build',
        'ci',
        'chore',
        'revert',
        'major',
        'minor',
        'patch',
        'deps',
      ],
    ],
  },
};
