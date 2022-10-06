package client

import "sync"

type walletManager struct {
	wallets     map[string]*Wallet
	walletMutex sync.Mutex
}

var globalWalletManager = &walletManager{
	wallets: make(map[string]*Wallet),
}

func (w *walletManager) isExistingWallet(privateKey string) (*Wallet, bool) {
	w.walletMutex.Lock()
	defer w.walletMutex.Unlock()
	wallet, exists := w.wallets[privateKey]
	return wallet, exists
}

func (w *walletManager) addNewWallet(privateKey string, newWallet *Wallet) {
	w.walletMutex.Lock()
	defer w.walletMutex.Unlock()
	w.wallets[privateKey] = newWallet
}
