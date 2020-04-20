<!--
Copyright 2019 F5 Networks. All rights reserved.
Use of this source code is governed by a MIT-style
license that can be found in the LICENSE file.
-->

# Contributing Guide for go-reducers

If you have found this that means you you want to help us out. Thanks in advance for lending a hand! This guide should get you up and running quickly and make it easy for you to contribute.  

## Issues

We use issues for bug reports and to discuss new features. If you are planning on contributing a new feature, you should open an issue first that discusses the feature you're adding. This will avoid wasting your time if someone else is already working on it, or if there's some design change that we need.

Creating issues is good, creating good issues is even better. Filing meaningful bug reports with lots of information in them helps us figure out what to fix when and how it impacts our users. We like bugs because it means people are using our code, and we like fixing them even more.

Please follow these guidelines for filing issues.

* Describe the problem
* Include detailed information about how to recreate the issue
* Include example code that illustrates the problem.  Better yet write a unit test.

## Pull Requests

We use [circle-ci](https://circleci.com/gh/weberr13/go-reducers) to automatically run some hooks on every pull request. These must pass before a pull request is accepted. You can see the details in [.circleci/config.yml](https://github.com/weberr13/go-reducers/blob/master/.circleci/config.yml). If your pull request fails in circleCI, your pull request will be blocked with a link to the failing run. Generally, we run these hooks:

* Unit tests are executed
* Code coverage data is collected
* Formatting and linting checks are performed

If you are submitting a pull request, you need to make sure that you have done a few things first.

* Make sure you have tested your code. Reviewers usually expect new unit tests for new code.
* The master branch must be kept release ready at all times. This requires that a single pull request should contain the code changes, unit tests, and documentation changes (if any)
* Use proper formatting for the code
* Clean up your git history because no one wants to see 75 commits for one issue
* Use the commit message format shown below

## Commit Message Format

The commit message for your final commit should use the following format:
```
Fix #<issue-number>: <One line summarizing what changed>

Problem: Brief description of the problem.

Solution: Detailed description of your solution

Testing (optional if not described in Solution section): Description of tests that were run to exercise the solutions (unit tests, system tests, etc)

affects-branches: branch1, branch2
```

* The messages should be line-wrapped to 80 characters if possible. Try to keep the one line summary under 80 characters if possible.
* If a commit fixes many issues, list all of them
* A line stating what branches the pull request is going to be merged into is required. The note should follow the format "affects-branches: branch1, branch2". This is because we have a robot that can check if bugfixes have been appropriately backported. This is only needed for bugfixes, and if you don't know what to put here for a bug, ask in your pull request.

## Testing

Creating tests is pretty straight forward and we need you to help us ensure the quality of our code. Every public API should have associated unit tests. We use the [GoConvey](https://github.com/smartystreets/goconvey/) BDD testing framework for Go.

## License

See the LICENSE file for our MIT style license

### Contributor License Agreement

Individuals or business entities who contribute to this project must have completed and submitted the [F5® Contributor License Agreement](http://clouddocs.f5.com/tbd.html) to TBD@f5.com prior to their code submission being included in this project. Please include your github handle in the CLA email.