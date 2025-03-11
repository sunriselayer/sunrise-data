package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/sunriselayer/sunrise-data/config"
	appctx "github.com/sunriselayer/sunrise-data/context"
	"github.com/sunriselayer/sunrise-data/optimism"
	"github.com/sunriselayer/sunrise-data/protocols"
)

/**
 * @description Optimismコマンドを実装するモジュール
 * Optimism DAサーバーを起動するためのコマンドを提供します
 */

// optimismCmd はOptimism DAサーバーを起動するためのコマンドです
var optimismCmd = &cobra.Command{
	Use:   "optimism",
	Short: "Optimism DAサーバーを起動",
	Long:  `このコマンドはOptimism DAサーバーを起動します。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 設定ファイルの読み込み
		config, err := config.LoadConfig()
		if err != nil {
			log.Error().Msgf("設定ファイルの読み込みに失敗しました: %s", err)
			return err
		}

		// 公開コンテキストの取得
		if err = appctx.GetPublishContext(*config); err != nil {
			log.Error().Msgf("sunrised RPCへの接続に失敗しました: %s", err)
			return err
		}

		// IPFSの接続確認
		if err := protocols.CheckIpfsConnection(); err != nil {
			log.Error().Msgf("IPFSへの接続に失敗しました: %s", err)
			return err
		}

		// Optimism DAサーバーの起動
		log.Info().Msg("Optimism DAサーバーを起動しています...")
		optimism.Start()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(optimismCmd)
}
