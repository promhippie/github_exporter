# Contributing to GitHub Exporter

Welcome! Our community focuses on helping others and making GitHub Exporter the best it can be. We gladly accept contributions and encourage you to get involved!


## Bug reports

Please search the issues on the issue tracker with a variety of keywords to ensure your bug is not already reported.

If unique, [open an issue](https://github.com/webhippie/github_exporter/issues) and answer the questions so we can understand and reproduce the problematic behavior.

The burden is on you to convince us that it is actually a bug in GitHub Exporter. This is easiest to do when you write clear, concise instructions so we can reproduce the behavior (even if it seems obvious). The more detailed and specific you are, the faster we will be able to help you. Check out [How to Report Bugs Effectively](http://www.chiark.greenend.org.uk/~sgtatham/bugs.html).

Please be kind, remember that GitHub Exporter comes at no cost to you, and you're getting free help.


## Check for assigned people

We are using Github Issues for submitting known issues (e.g. bugs, features, etc.). Some issues will have someone assigned, meaning that there's already someone that takes responsability for fixing said issue. This is not done to discourage contributions, rather to not step in the work that has already been done by the assignee. If you want to work on a known issue with someone already assigned to it, please consider contacting the assignee first (e.g. by mentioning the assignee in a new comment on the specific issue). This way you can contribute with ideas, or even with code if the assignee decides that you can step in.

If you plan to work on a non assigned issue, please add a comment on the issue to prevent duplicated work.


## Minor improvements and new tests

Submit pull requests at any time for minor changes or new tests. Make sure to write tests to assert your change is working properly and is thoroughly covered. We'll ask most pull requests to be squashed, especially with small commits.

Your pull request may be thoroughly reviewed. This is because if we accept the PR, we also assume responsibility for it, although we would prefer you to help maintain your code after it gets merged.


## Mind the Style

We believe that in order to have a healthy codebase we need to abide to a certain code style. We use `gofmt` with Go and `eslint` with Javscript for this matter, which are tools that has proved to be useful. So, before submitting your Pull Request, make sure that `gofmt` and if viable `eslint` are passing for you.

Finally, note that `gofmt` and if viable `eslint` are called on the CI system. This means that your Pull Request will not be merged until the changes are approved.


## Update the Changelog

We keep a changelog in the `CHANGELOG.md` file. This is useful to understand what has changed between each version. When you implement a new feature, or a fix for an issue, please also update the `CHANGELOG.md` file accordingly. We don't follow a strict style for the changelog, just try to be consistent with the rest of the file.


## Sign your work

The sign-off is a simple line at the end of the explanation for the patch. Your signature certifies that you wrote the patch or otherwise have the right to pass it on as an open-source patch. The rules are pretty simple: If you can certify [DCO](DCO), then you just add a line to every git commit message:

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

Please use your real name, we really dislike pseudonyms or anonymous contributions. We are in the opensource world without secrets. If you set your `user.name` and `user.email` git configs, you can sign your commit automatically with `git commit -s`.


## Collaborator status

If your pull request is merged, congratulations! You're technically a collaborator. We may also grant you "Collaborator status" which means you can push to the repository and merge other pull requests. We hope that you will stay involved by reviewing pull requests, submitting more of your own, and resolving issues as you are able to. Thanks for making GitHub Exporter amazing!

We ask that collaborators will conduct thorough code reviews and be nice to new contributors. Before merging a PR, it's best to get the approval of at least one or two other collaborators and/or the project owner. We prefer squashed commits instead of many little, semantically-unimportant commits. Also, CI and other post-commit hooks must pass before being merged except in certain unusual circumstances.

Collaborator status may be removed for inactive users from time to time as we see fit; this is not an insult, just a basic security precaution in case the account becomes inactive or abandoned. Privileges can always be restored later.

**Reviewing pull requests:** Please help submit and review pull requests as you are able! We would ask that every pull request be reviewed by at least one collaborator who did not open the pull request before merging. This will help ensure high code quality as new collaborators are added to the project.


## Vulnerabilities

If you've found a vulnerability that is serious, please email to thomas@webhippie.de. If it's not a big deal, a pull request will probably be faster.


## Thank you

Thanks for your help! GitHub Exporter would not be what it is today without your contributions.
