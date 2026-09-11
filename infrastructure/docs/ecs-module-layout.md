# ECR・ECSのモジュール構成

cloud-pratica-terraformの分け方を参考に、環境側からECR・ECS・タスク定義をそれぞれ1つのモジュールとして呼び出す。今回のプロジェクトはAWS専用のため、modules/aws階層は追加しない。

`````text
infrastructure/
├── stg/
│   ├── aws.tf                 # 各モジュールの呼び出し・依存関係
│   └── moved.tf               # 旧アドレスからの移動定義
└── modules/
    ├── ecr/
    │   └── main.tf            # backend・migrationをecr_unitで作成
    ├── ecr_unit/
    │   └── main.tf            # リポジトリ1個とライフサイクル
    ├── ecs/
    │   └── main.tf            # ECSクラスター（サービスは未実装）
    └── ecs_task_definition/
        ├── main.tf            # バックエンド・マイグレーションの2リソース
        └── variables.tf       # backend・migrationのobject型変数
`````

各モジュールの入力はvariables.tf、他モジュールへ渡す値はoutputs.tfで管理する。IAM・ログ・SGは既存の専用モジュールに残す。参考先のFARGATE_SPOTやECS Exec設定などはコピーせず、今回はディレクトリと責務の分け方だけを合わせた。

タスク定義はmain.tfに集約する。ポート公開・Secret・起動コマンドが異なるため、aws_ecs_task_definitionは2つに分ける。入力はbackendとmigrationのobject型でまとめ、region・tagsだけを共通で渡す。リソースのアドレスと設定値はこの整理では変更しない。

AWS上のリソース名・イメージ・スペック・権限は変更しない。Terraformのアドレス変更にはmovedブロックを用意し、既存ECR等を削除・再作成しない。移動は次のapplyでStateに反映される。未作成のリソースについては通常どおり作成される。
