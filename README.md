# Avapmost

**Avapmost** は [Mattermost](https://mattermost.com) v11 をベースにした、AVAP Co., Ltd. によるフォークです。
Mattermost v10 → v11 のアップデートでフリーティア (Team Edition) から削除された機能を復元し、業務利用を継続できるよう公開しています。

> **このリポジトリが役立つ方**
> Mattermost を無償で自社運用していて、v11 へのアップデートで以下の機能が使えなくなり困っている組織・個人を主な対象としています。

---

## Mattermost v11 からの変更点

| 機能 | Mattermost v11 Team Edition | Avapmost |
|---|---|---|
| ユーザー数上限 | 有料プランへの誘導 UI あり・上限強制 | 上限撤廃（UI も除去） |
| CJK 全文検索（日本語・中国語・韓国語） | 削除（Bleve による無料検索が廃止） | 復元（CJKSearch フラグをデフォルト有効化） |
| GitLab SSO | Professional 以上のみ | 復元（Team Edition でも利用可） |
| 検索プラグイン (avapmost-search) | なし | 同梱（Elasticsearch + セマンティック検索対応） |

その他のリブランディング変更：
- 製品名・ロゴ・アイコンを Avapmost に変更（商標ガイドライン準拠）
- エンタープライズへのアップセル UI を非表示化

---

## ライセンスと適法性

本リポジトリは Mattermost のライセンス条項に従って作成されています。

### 準拠したライセンス条項

Mattermost のソースコードには複数のライセンスが適用されており、本フォークは各条項を遵守しています。

**サーバー側 Go コード (`server/` 配下の大部分)**
[GNU AGPL v3.0](LICENSE.txt) に基づきます。AGPL はフォーク・改変・再配布を認めており、その際にソースコードを公開する義務があります。本リポジトリの公開がその義務の履行に相当します。

**フロントエンド コード (`webapp/` 配下)**
[Apache License v2.0](LICENSE.txt) に基づきます（LICENSE.txt 内の Mattermost Licensing ポリシーにて `webapp/ and all subdirectories thereof` が Apache v2.0 の対象と明記されています）。Apache v2.0 は改変・再配布を許可しています。

**エンタープライズコード (`server/enterprise/` 配下)**
Mattermost Source Available License（有償ライセンス要）の対象です。**本フォークはこのディレクトリを一切変更していません。**

**商標**
"Mattermost" の商標は Mattermost, Inc. に帰属します。商標ガイドライン（[Trademark Standards of Use](https://mattermost.com/trademark-standards-of-use/)）に従い、製品名・ロゴを "Avapmost" に変更しています。ログインフッターおよび About ダイアログの Mattermost, Inc. 著作権表示は原文のまま保持しています。

### 免責事項

本リポジトリは現状有姿で提供されます。セキュリティパッチの適用は利用者の責任において行ってください。Mattermost, Inc. のサポート対象外となります。

---

## Docker イメージ

```bash
docker pull avap.plus/public/avapmost:latest
```

バージョン履歴は [VERSIONS.md](VERSIONS.md) を参照してください。

---

## ビルド方法

```bash
# 開発環境の起動
make dev-start

# Docker イメージのビルド
make image-build-fast

# イメージのビルド & push
make image-push TAG=11.5.0-avap.1
```

詳細は [Makefile](Makefile) を参照してください。

---

*以下はオリジナルの Mattermost README です。*

---

# [![Mattermost logo](https://user-images.githubusercontent.com/7205829/137170381-fe86eef0-bccc-4fdd-8e92-b258884ebdd7.png)](https://mattermost.com)

[Mattermost](https://mattermost.com) is an open core, self-hosted collaboration platform that offers chat, workflow automation, voice calling, screen sharing, and AI integration. This repo is the primary source for core development on the Mattermost platform; it's written in Go and React, runs as a single Linux binary, and relies on PostgreSQL. A new compiled version is released under an MIT license every month on the 16th.

[Deploy Mattermost on-premises](https://mattermost.com/deploy/?utm_source=github-mattermost-server-readme), or [try it for free in the cloud](https://mattermost.com/sign-up/?utm_source=github-mattermost-server-readme).

<img width="1006" alt="mattermost user interface" src="https://user-images.githubusercontent.com/7205829/136107976-7a894c9e-290a-490d-8501-e5fdbfc3785a.png">

Learn more about the following use cases with Mattermost:

- [DevSecOps](https://mattermost.com/solutions/use-cases/devops/?utm_source=github-mattermost-server-readme)
- [Incident Resolution](https://mattermost.com/solutions/use-cases/incident-resolution/?utm_source=github-mattermost-server-readme)
- [IT Service Desk](https://mattermost.com/solutions/use-cases/it-service-desk/?utm_source=github-mattermost-server-readme)

Other useful resources:

- [Download and Install Mattermost](https://docs.mattermost.com/guides/deployment.html) - Install, setup, and configure your own Mattermost instance.
- [Product documentation](https://docs.mattermost.com/) - Learn how to run a Mattermost instance and take advantage of all the features.
- [Developer documentation](https://developers.mattermost.com/) - Contribute code to Mattermost or build an integration via APIs, Webhooks, slash commands, Apps, and plugins.

Table of contents
=================

- [Install Mattermost](#install-mattermost)
- [Native mobile and desktop apps](#native-mobile-and-desktop-apps)
- [Get security bulletins](#get-security-bulletins)
- [Get involved](#get-involved)
- [Learn more](#learn-more)
- [License](#license)
- [Get the latest news](#get-the-latest-news)
- [Contributing](#contributing)

## Install Mattermost

- [Download and Install Mattermost Self-Hosted](https://docs.mattermost.com/guides/deployment.html) - Deploy a Mattermost Self-hosted instance in minutes via Docker, Ubuntu, or tar.
- [Get started in the cloud](https://mattermost.com/sign-up/?utm_source=github-mattermost-server-readme) to try Mattermost today.
- [Developer machine setup](https://developers.mattermost.com/contribute/server/developer-setup) - Follow this guide if you want to write code for Mattermost.


Other install guides:

- [Deploy Mattermost on Docker](https://docs.mattermost.com/install/install-docker.html)
- [Mattermost Omnibus](https://docs.mattermost.com/install/installing-mattermost-omnibus.html)
- [Install Mattermost from Tar](https://docs.mattermost.com/install/install-tar.html)
- [Ubuntu 20.04 LTS](https://docs.mattermost.com/install/installing-ubuntu-2004-LTS.html)
- [Kubernetes](https://docs.mattermost.com/install/install-kubernetes.html)
- [Helm](https://docs.mattermost.com/install/install-kubernetes.html#installing-the-operators-via-helm)
- [Debian Buster](https://docs.mattermost.com/install/install-debian.html)
- [RHEL 8](https://docs.mattermost.com/install/install-rhel-8.html)
- [More server install guides](https://docs.mattermost.com/guides/deployment.html)

## Native mobile and desktop apps

In addition to the web interface, you can also download Mattermost clients for [Android](https://mattermost.com/pl/android-app/), [iOS](https://mattermost.com/pl/ios-app/), [Windows PC](https://docs.mattermost.com/install/desktop-app-install.html#windows-10-windows-8-1), [macOS](https://docs.mattermost.com/install/desktop-app-install.html#macos-10-9), and [Linux](https://docs.mattermost.com/install/desktop-app-install.html#linux).

[<img src="https://user-images.githubusercontent.com/30978331/272826427-6200c98f-7319-42c3-86d4-0b33ae99e01a.png" alt="Get Mattermost on Google Play" height="50px"/>](https://mattermost.com/pl/android-app/)  [<img src="https://developer.apple.com/assets/elements/badges/download-on-the-app-store.svg" alt="Get Mattermost on the App Store" height="50px"/>](https://itunes.apple.com/us/app/mattermost/id1257222717?mt=8)  [![Get Mattermost on Windows PC](https://user-images.githubusercontent.com/33878967/33095357-39cab8d2-ceb8-11e7-89a6-67dccc571ca3.png)](https://docs.mattermost.com/install/desktop.html#windows-10-windows-8-1-windows-7)  [![Get Mattermost on Mac OSX](https://user-images.githubusercontent.com/33878967/33095355-39a36f2a-ceb8-11e7-9b33-73d4f6d5d6c1.png)](https://docs.mattermost.com/install/desktop.html#macos-10-9)  [![Get Mattermost on Linux](https://user-images.githubusercontent.com/33878967/33095354-3990e256-ceb8-11e7-965d-b00a16e578de.png)](https://docs.mattermost.com/install/desktop.html#linux)

## Get security bulletins

Receive notifications of critical security updates. The sophistication of online attackers is perpetually increasing. If you're deploying Mattermost it's highly recommended you subscribe to the Mattermost Security Bulletin mailing list for updates on critical security releases.

[Subscribe here](https://mattermost.com/security-updates/#sign-up)

## Get involved

- [Contribute to Mattermost](https://handbook.mattermost.com/contributors/contributors/ways-to-contribute)
- [Find "Help Wanted" projects](https://github.com/mattermost/mattermost-server/issues?page=1&q=is%3Aissue+is%3Aopen+%22Help+Wanted%22&utf8=%E2%9C%93)
- [Join Developer Discussion on a Mattermost server for contributors](https://community.mattermost.com/signup_user_complete/?id=f1924a8db44ff3bb41c96424cdc20676)
- [Get Help With Mattermost](https://docs.mattermost.com/guides/get-help.html)

## Learn more

- [API options - webhooks, slash commands, drivers, and web service](https://api.mattermost.com/)
- [See who's using Mattermost](https://mattermost.com/customers/)
- [Browse over 700 Mattermost integrations](https://mattermost.com/marketplace/)

## License

See the [LICENSE file](LICENSE.txt) for license rights and limitations.

## Get the latest news

- **X** - Follow [Mattermost on X, formerly Twitter](https://twitter.com/mattermost).
- **Blog** - Get the latest updates from the [Mattermost blog](https://mattermost.com/blog/).
- **Facebook** - Follow [Mattermost on Facebook](https://www.facebook.com/MattermostHQ).
- **LinkedIn** - Follow [Mattermost on LinkedIn](https://www.linkedin.com/company/mattermost/).
- **Email** - Subscribe to our [newsletter](https://mattermost.us11.list-manage.com/subscribe?u=6cdba22349ae374e188e7ab8e&id=2add1c8034) (1 or 2 per month).
- **Mattermost** - Join the ~contributors channel on [the Mattermost Community Server](https://community.mattermost.com).
- **IRC** - Join the #matterbridge channel on [Freenode](https://freenode.net/) (thanks to [matterircd](https://github.com/42wim/matterircd)).
- **YouTube** -  Subscribe to [Mattermost](https://www.youtube.com/@MattermostHQ).

## Contributing

[![Small Image](https://img.shields.io/badge/Contribute%20with-Gitpod-908a85?logo=gitpod)](https://gitpod.io/#https://github.com/mattermost/mattermost)

Please see [CONTRIBUTING.md](./CONTRIBUTING.md).
[Join the Mattermost Contributors server](https://community.mattermost.com/signup_user_complete/?id=codoy5s743rq5mk18i7u5ksz7e) to join community discussions about contributions, development, and more.
