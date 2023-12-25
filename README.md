
# こちらはgoとposgresql(sqlc)を用いた簡単なappになります。

## base

- 基本は全てmakeファイルに書いてあります。
- tableを追加する場合は、dbdiagramからテーブルを追加&export -> migrationファイルを作成 -> dbをup -> queryファイルに記載してファイルを作成的な流れです。
- とってもわかりやすいですね。

> [!WARNING]
> DB周りで型定義エラーが起きています。動作検証できていない。。


## 参考

[Backend master input](https://www.youtube.com/playlist?list=PLy_6D98if3ULEtXtNSY_2qN21VCKgoQAE)
