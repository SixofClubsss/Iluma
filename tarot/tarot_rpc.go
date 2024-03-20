package tarot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/dReam-dApps/dReams/rpc"
	dero "github.com/deroproject/derohe/rpc"
)

// Get Tarot SC data
func FetchTarotSC() {
	if rpc.Daemon.IsConnected() && rpc.Wallet.Height() > Iluma.Value.Height {
		client, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.TarotSCID,
			Code:      false,
			Variables: true,
		}

		if err := client.CallFor(ctx, &result, "DERO.GetSC", params); err != nil {
			logger.Errorln("[FetchTarotSC]", err)
			return
		}

		Iluma.Value.Height = rpc.Wallet.Height()

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			Iluma.Value.Readings = fmt.Sprint(Reading_jv)
		}
	}
}

// Find Tarot reading on SC
func FetchReading(tx string) {
	if rpc.Daemon.IsConnected() && len(tx) == 64 {
		client, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.TarotSCID,
			Code:      false,
			Variables: true,
		}

		err := client.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			logger.Errorln("[FetchTarotReading]", err)
			return
		}

		Reading_jv := result.VariableStringKeys["readings:"]
		if Reading_jv != nil {
			Display_jv := result.VariableStringKeys["Display"]
			start := rpc.IntType(Reading_jv) - rpc.IntType(Display_jv)
			i := start
			for i < start+45 {
				h := "-readingTXID:"
				w := strconv.Itoa(i)
				TXID_jv := result.VariableStringKeys[w+h]

				if TXID_jv != nil {
					if fmt.Sprint(TXID_jv) == tx {
						Iluma.Value.Found = true
						Iluma.Value.Card1 = findTarotCard(result.VariableStringKeys[w+"-card1:"])
						Iluma.Value.Card2 = findTarotCard(result.VariableStringKeys[w+"-card2:"])
						Iluma.Value.Card3 = findTarotCard(result.VariableStringKeys[w+"-card3:"])
					}
				}
				i++
			}
		}
	}
}

// Find Tarot card from hash value
func findTarotCard(hash interface{}) int {
	if hash != nil {
		for i := 1; i < 79; i++ {
			finder := strconv.Itoa(i)
			card := sha256.Sum256([]byte(finder))
			str := hex.EncodeToString(card[:])

			if str == fmt.Sprint(hash) {
				return i
			}
		}
	}
	return 0
}

// Draw Iluma Tarot reading from SC
//   - num defines one or three card draw
func DrawReading(num int) (tx string) {
	args := dero.Arguments{
		dero.Argument{Name: "entrypoint", DataType: "S", Value: "Draw"},
		dero.Argument{Name: "num", DataType: "U", Value: uint64(num)},
	}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        rpc.IlumaFee,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.TarotSCID, "[Iluma]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.TarotSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpc.Wallet.CallFor(&txid, "transfer", params); err != nil {
		rpc.PrintError("[Iluma] Tarot Reading: %s", err)
		return
	}

	Iluma.Value.Num = num
	Iluma.Value.Last = txid.TXID
	Iluma.Value.Notified = false

	rpc.PrintLog("[Iluma] Tarot Reading TX: %s", txid)

	Iluma.Value.CHeight = rpc.Wallet.Height()

	return txid.TXID
}
