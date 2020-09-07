<p align="center">
  <img src="https://user-images.githubusercontent.com/9434884/43697219-3cb4ef3a-9975-11e8-9a9c-73f4f537442d.png" alt="Sentinel Logo" width="50%"/>
</p>
<h2 align="center">
Sentinel: The Sentinel of Your Microservices
</h2>

<p align="center">
  <a>
    <img alt="CI" src="https://github.com/alibaba/sentinel-golang/workflows/CI/badge.svg?branch=master"></a>
  <a href="https://travis-ci.org/alibaba/sentinel-golang">
    <img alt="Build Status" src="https://travis-ci.org/alibaba/sentinel-golang.svg?branch=master"></a>
  <a href="https://codecov.io/gh/alibaba/sentinel-golang">
    <img alt="codecov" src="https://codecov.io/gh/alibaba/sentinel-golang/branch/master/graph/badge.svg"></a>
  <a href="tps://goreportcard.com/report/github.com/alibaba/sentinel-golang">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/alibaba/sentinel-golang"></a>
  <a href="https://www.apache.org/licenses/LICENSE-2.0.html">
    <img alt="License" src="https://img.shields.io/badge/license-Apache%202-4EB1BA.svg"></a>
  <a href="https://gitter.im/alibaba/Sentinel">
    <img alt="Gitter" src="https://badges.gitter.im/alibaba/Sentinel.svg"></a>
  <a href="https://github.com/alibaba/sentinel-golang/commits/dev">
    <img alt="GitHub last commit" src="https://img.shields.io/github/last-commit/alibaba/sentinel-golang.svg?style=flat-square"></a>
  <a href="https://github.com/alibaba/sentinel-golang">
    <img alt="GitHub repo size" src="https://img.shields.io/github/repo-size/alibaba/sentinel-golang"></a>
  <a href="alibaba/sentinel-golang/issues?q=is%3Aissue+is%3Aclosed">
    <img alt = "GitHub closed issues" src="https://img.shields.io/github/issues-closed/alibaba/sentinel-golang.svg?style=flat-square"></a>
</p>

## Introduction

As distributed systems become increasingly popular, the reliability between services is becoming more important than ever before.
Sentinel takes "flow" as breakthrough point, and works on multiple fields including **flow control**,
**circuit breaking** and **system adaptive protection**, to guarantee reliability and resiliency of microservices.

![flow-overview](https://raw.githubusercontent.com/sentinel-group/sentinel-website/master/img/sentinel-flow-index-overview-en.jpg)

Sentinel provides the following features:

- **Rich applicable scenarios**: Sentinel has been wildly used in Alibaba, and has covered almost all the core-scenarios in Double-11 (11.11) Shopping Festivals in the past 10 years, such as “Second Kill” which needs to limit burst flow traffic to meet the system capacity, throttling, circuit breaking for unreliable downstream services, distributed rate limiting, etc.
- **Real-time monitoring**: Sentinel also provides real-time monitoring ability. You can see the runtime information of a single machine in real-time, and pump the metrics to outside metric components like Prometheus.
- **Cloud-native ecosystem**: Sentinel Go provides out-of-box integrations with cloud-native components.

## Documentation

See the [中文文档](https://github.com/alibaba/sentinel-golang/wiki/介绍) for document in Chinese.

See the [Wiki](https://github.com/alibaba/sentinel-golang/wiki) for full documentation, examples, blog posts, and other information.

If you are using Sentinel, please [**leave a comment here**](https://github.com/alibaba/Sentinel/issues/18) to tell us your scenario to make Sentinel better.
It's also encouraged to add the link of your blog post, tutorial, demo or customized components to [**Awesome Sentinel**](https://github.com/alibaba/sentinel-awesome).

## Bugs and Feedback

For bug report, questions and discussions please submit [GitHub Issues](https://github.com/alibaba/sentinel-golang/issues).

## Contributing

Contributions are always welcomed! Please see [CONTRIBUTING](./CONTRIBUTING.md) for detailed guidelines.

You can start with the issues labeled with [`good first issue`](https://github.com/alibaba/sentinel-golang/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).
