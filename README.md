<p align="center">
<img src="https://i.imgur.com/d9UWkck.png" alt="logo"></a>
</p>

webhook-bridge
==============
Bridge Webhook into your tool or internal integration.
Like:
[Sentry](https://sentry.io)
[WeChat](https://www.wechat.com/en/),
[POPO](http://popo.netease.com/)

[![Python Version](https://img.shields.io/pypi/pyversions/webhook-bridge)](https://img.shields.io/pypi/pyversions/webhook-bridge)
[![Nox](https://img.shields.io/badge/%F0%9F%A6%8A-Nox-D85E00.svg)](https://github.com/wntrblm/nox)
[![PyPI Version](https://img.shields.io/pypi/v/webhook-bridge?color=green)](https://pypi.org/project/webhook-bridge/)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge)](https://pepy.tech/project/webhook-bridge)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge/month)](https://pepy.tech/project/webhook-bridge)
[![Downloads](https://static.pepy.tech/badge/webhook-bridge/week)](https://pepy.tech/project/webhook-bridge)
[![License](https://img.shields.io/pypi/l/webhook-bridge)](https://pypi.org/project/webhook-bridge/)
[![PyPI Format](https://img.shields.io/pypi/format/webhook-bridge)](https://pypi.org/project/webhook-bridge/)
[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://github.com/loonghao/webhook-bridge/graphs/commit-activity)


<p align="center">
<img src="https://i.imgur.com/31RO4xN.png" alt="logo"></a>
</p>


# Setup Dev Environment

```shell
pip install -r requirements-dev.txt
```

## Test in local
This command line will auto load example plugins.
```shell
nox -s local-test
```

Test post data to the webhook bridge
```shell script
curl -X POST "http://localhost:5001/api/plugin/sentry" -H  "accept: application/json" -H  "Content-Type: application/json" -d "[[\"browser\",\"Chrome 28.0.1500\"],[\"browser.name\",\"Chrome\"],[\"client_os\",\"Windows 8\"],[\"client_os.name\",\"Windows\"],[\"environment\",\"prod\"],[\"level\",\"error\"],[\"sentry:user\",\"id:1\"],[\"server_name\",\"web01.example.org\"],[\"url\",\"http://example.com/foo\"]]"
```
If everything is set up properly, you will see that the plugin is executed normally.

<img src="https://i.imgur.com/QnVVdor.gif" alt="logo"></a>


# Installing

You can install via pip.

```cmd
pip install webhook-bridge
```
