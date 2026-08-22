package utils

import (
	"Wallet-App/model"
	"testing"
)

func Compare_transaction_structs(t *testing.T, stage string, received *model.Transaction, expected *model.Transaction) {
	//first compare pointers.
	if received == nil || expected == nil {
		if received == nil && expected != nil {
			t.Errorf("Boundary (%s): expected transaction struct %+v got nil pointer", stage, expected)
		} else if received != nil && expected == nil {
			t.Errorf("Boundary (%s): expected nil transaction struct got %+v", stage, received)
		}
	} else {
		//compare received struct fields to the expected struct fields.
		if received.WalletID != expected.WalletID {
			t.Errorf("Logic (%s): expected WalletID %d got %d", stage, expected.WalletID, received.WalletID)
		}
		if received.Type != expected.Type {
			t.Errorf("Logic (%s): expected type %s got %s", stage, expected.Type, received.Type)
		}
		if received.Amount != expected.Amount {
			t.Errorf("Logic (%s): expected Amount %d got %d", stage, expected.Amount, received.Amount)
		}
		if received.Category != expected.Category {
			t.Errorf("Logic (%s): expected category %s got %s", stage, expected.Category, received.Category)
		}
		//check the related wallet ID.
		if received.RelatedWalletID == nil || expected.RelatedWalletID == nil {
			if received.RelatedWalletID == nil && expected.RelatedWalletID != nil {
				t.Errorf("Boundary (%s): expected not nil related wallet id got %v", stage, received.RelatedWalletID)
			} else if received.RelatedWalletID != nil && expected.RelatedWalletID == nil {
				t.Errorf("Boundary (%s): expected nil related wallet ID got %v", stage, received.RelatedWalletID)
			}
		} else {
			//compare the id in expected transaction with actual one.
			if *received.RelatedWalletID != *expected.RelatedWalletID {
				t.Errorf("Logic (%s): expected related wallet ID in transaction %d got %d", stage, *expected.RelatedWalletID, *received.RelatedWalletID)
			}
		}
	}
}
