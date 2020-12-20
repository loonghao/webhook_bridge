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


<p align="center">
<img src="https://i.imgur.com/31RO4xN.png" alt="logo"></a>
</p>

Installing
----------
You can install via pip.

```cmd
pip install webhook_bridge
```
or through clone from Github.
```git exclude
git clone https://github.com/loonghao/webhook_bridge.git
```
Install package.
```cmd
python setup.py install
```

QuickStart
----------
# Launch server.
```shell script
# Load example plugin for test.
set WEBHOOK_BRIDGE_SERVER_PLUGINS=C:\Users\hao.long\webhook_bridge_server\example_plugins
webhook-bridge
```
Test post data to the webhook bridge
```shell script
curl -X POST "http://localhost:5001/api/plugin/sentry" -H  "accept: application/json" -H  "Content-Type: application/json" -d "[[\"browser\",\"Chrome 28.0.1500\"],[\"browser.name\",\"Chrome\"],[\"client_os\",\"Windows 8\"],[\"client_os.name\",\"Windows\"],[\"environment\",\"prod\"],[\"level\",\"error\"],[\"sentry:user\",\"id:1\"],[\"server_name\",\"web01.example.org\"],[\"url\",\"http://example.com/foo\"]]"
```
If everything is set up properly, you will see that the plugin is executed normally.

<img src="https://i.imgur.com/QnVVdor.gif" alt="logo"></a>

local docs power by fastapi
`http://localhost:5001/docs`
