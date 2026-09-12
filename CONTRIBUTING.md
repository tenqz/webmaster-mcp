# Contributing Guide

Thank you for contributing to Yandex Webmaster MCP. This guide outlines quality expectations and the contribution process.

## Quick Start

**Requirements:** Go 1.25+

```bash
make check
```

Run container integration checks with `make docker-test`. Tests use fixtures and never require Yandex credentials.

## Commits — Iterative and Atomic

Follow [ACDD](https://opatsay.com/ru/atomarnye-kommity-acdd):

- Every commit changes exactly one file, including tests, dependencies, configuration and permissions.
- Keep each intermediate commit buildable; split cross-file work into compatible steps.
- Commit regression tests separately immediately after the verified fix.
- Never squash a multi-file pull request into one commit. Preserve the single-file sequence.
- Verify each step before committing; review every commit before merging.

## Mandatory Comments and Tests

- Public types and non-trivial logic must include comments that explain the **why**
- New features **must include unit tests**; bug fixes should include a regression test
- Do not regress quality: run `make check` before pushing

## Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
<type>[optional scope]: <imperative short summary>

[optional body]
```

| Type | Description |
| --- | --- |
| `feat` | New functionality |
| `fix` | Bug fixes |
| `refactor` | Refactoring without changing behavior |
| `perf` | Performance improvements |
| `test` | Add or fix tests |
| `chore` | Technical changes that do not change behavior |
| `docs` | Documentation |
| `style` | Formatting |
| `build` | Build system and dependencies |
| `ci` | CI/CD configuration |

### Examples

```text
feat(webmaster): add Yandex Webmaster domain types
test(webmaster): cover analytics date validation
docs: describe remote MCP agent configuration
ci: run unit tests on push to main
```

## Branch Naming

- `feat/` — new features
- `fix/` — bug fixes
- `docs/` — documentation only
- `ci/` — CI configuration

## Security

- Never commit secrets, keys, or `.env` files
- Report security issues privately to: smmartbiz@gmail.com

## Documentation

- Update **README.md** for user-facing changes
- Keep godoc on exported types current

## Getting Help

- **Bugs:** open an issue with reproduction steps
- **Features:** open an issue before implementing large changes
- **Security:** email smmartbiz@gmail.com privately

## Code of Conduct

- Be respectful and constructive
- Welcome newcomers
- Focus on what is best for the community

Thank you for your contribution.
