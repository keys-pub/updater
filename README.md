# Updater

## Check for Update

```shell
updater -github keys-pub/app -app-name Keys -current 0.0.17

{
  "version": "0.0.18",
  "publishedAt": 1583275443689,
  "asset": {
    "name": "Keys-0.0.18-mac.zip",
    "url": "https://github.com/keys-pub/app/releases/latest/download/Keys-0.0.18-mac.zip",
    "digest": "9fe462603acbd84e55e5dfa6a02f40d0483551c88bd053b4b3827aba67d7fe3e53414a2214f6387a02e0bfc667d464ed0cc494f14b6ca04ae5ca81a20d503618",
    "digestType": "sha512",
    "localPath": ""
  },
  "needUpdate": true
}
```

## Download Update

```shell
updater -github keys-pub/app -app-name Keys -current 0.0.17 -download

{
  "version": "0.0.18",
  "publishedAt": 1583275443689,
  "asset": {
    "name": "Keys-0.0.18-mac.zip",
    "url": "https://github.com/keys-pub/app/releases/latest/download/Keys-0.0.18-mac.zip",
    "digest": "9fe462603acbd84e55e5dfa6a02f40d0483551c88bd053b4b3827aba67d7fe3e53414a2214f6387a02e0bfc667d464ed0cc494f14b6ca04ae5ca81a20d503618",
    "digestType": "sha512",
    "localPath": "/var/folders/84/7q50__3j4yscqtt4qmrv8rtr0000gn/T/updater/Keys/Keys-0.0.18-mac.zip"
  },
  "needUpdate": true
}
```

## Apply Update

```shell
updater -github keys-pub/app -app-name Keys -current 0.0.17 -download -apply /Applications/Keys.app
```
