const ids = {
  refreshStatus: document.querySelector("#refresh-status"),
  selectedWallet: document.querySelector("#selected-wallet"),
  walletForm: document.querySelector("#wallet-form"),
  walletSelect: document.querySelector("#wallet-select"),
  walletState: document.querySelector("#wallet-state"),
  walletBalance: document.querySelector("#wallet-balance"),
  walletName: document.querySelector("#wallet-name"),
  walletTxCount: document.querySelector("#wallet-txcount"),
  walletUtxos: document.querySelector("#wallet-utxos"),
  availableWallets: document.querySelector("#available-wallets"),
  loadedWallets: document.querySelector("#loaded-wallets"),
  sendWalletContext: document.querySelector("#send-wallet-context"),
  sendForm: document.querySelector("#send-form"),
  sendAddress: document.querySelector("#send-address"),
  sendAmount: document.querySelector("#send-amount"),
  sendState: document.querySelector("#send-state"),
  transactionList: document.querySelector("#transaction-list"),
};

const btcFormat = new Intl.NumberFormat("pt-BR", {
  maximumFractionDigits: 8,
});

const numberFormat = new Intl.NumberFormat("pt-BR");

async function fetchJSON(url, options = {}) {
  const response = await fetch(url, {
    headers: {
      Accept: "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `HTTP ${response.status}`);
  }

  return response.json();
}

function setText(element, value) {
  element.textContent = value;
}

function setStatus(message, isError = false) {
  ids.refreshStatus.classList.toggle("error", isError);
  setText(ids.refreshStatus, message);
}

function setWalletState(message, state) {
  ids.walletState.classList.remove("active", "error");

  if (state) {
    ids.walletState.classList.add(state);
  }

  setText(ids.walletState, message);
}

function setSendState(message, state) {
  ids.sendState.classList.remove("active", "error");

  if (state) {
    ids.sendState.classList.add(state);
  }

  setText(ids.sendState, message);
}

function renderList(list, wallets, emptyMessage) {
  list.innerHTML = "";

  if (!wallets || wallets.length === 0) {
    const item = document.createElement("li");
    item.className = "empty-state";
    item.textContent = emptyMessage;
    list.append(item);
    return;
  }

  for (const wallet of wallets) {
    const item = document.createElement("li");
    item.title = wallet;
    item.textContent = wallet;
    list.append(item);
  }
}

function renderOptions(wallets, selectedWallet) {
  ids.walletSelect.innerHTML = "";
  ids.walletSelect.disabled = wallets.length === 0;

  for (const wallet of wallets) {
    const option = document.createElement("option");
    option.value = wallet;
    option.textContent = wallet;
    option.selected = wallet === selectedWallet;
    ids.walletSelect.append(option);
  }
}

function updateWallets(data) {
  const availableWallets = data.available_wallets || [];
  const loadedWallets = data.loaded_wallets || [];
  const selectedWallet = data.selected_wallet || "";

  setText(ids.selectedWallet, selectedWallet || "--");
  ids.selectedWallet.title = selectedWallet;
  setText(ids.sendWalletContext, selectedWallet || "--");
  renderOptions(availableWallets, selectedWallet);
  renderList(ids.availableWallets, availableWallets, "Nenhuma wallet encontrada");
  renderList(ids.loadedWallets, loadedWallets, "Nenhuma wallet carregada");

  if (selectedWallet) {
    setWalletState(`Wallet selecionada: ${selectedWallet}`, "active");
  } else {
    setWalletState("Nenhuma wallet disponível");
  }
}

function updateWalletInfo(info) {
  setText(ids.walletName, info.walletname || "--");
  ids.walletName.title = info.walletname || "";
  setText(ids.walletBalance, `${btcFormat.format(info.balance || 0)} BTC`);
  setText(ids.walletTxCount, numberFormat.format(info.txcount || 0));
}

function updateWalletStatus(status) {
  if (!status || !status.wallet) {
    setText(ids.walletUtxos, "--");
    return;
  }

  setText(ids.walletName, status.wallet);
  ids.walletName.title = status.wallet;
  setText(ids.walletBalance, `${btcFormat.format(status.balance || 0)} BTC`);
  setText(ids.walletUtxos, numberFormat.format(status.utxos || 0));
  setText(ids.sendWalletContext, status.wallet);
}

function renderTransactions(transactions) {
  ids.transactionList.innerHTML = "";

  if (!transactions || transactions.length === 0) {
    const item = document.createElement("li");
    item.className = "empty-state";
    item.textContent = "Nenhuma transação enviada";
    ids.transactionList.append(item);
    return;
  }

  for (const transaction of transactions) {
    const item = document.createElement("li");
    const heading = document.createElement("div");
    const txid = document.createElement("span");
    const status = document.createElement("strong");
    const meta = document.createElement("p");
    const message = document.createElement("p");

    item.className = `transaction-item ${transaction.status || "unknown"}`;
    txid.className = "txid";
    txid.title = transaction.txid;
    txid.textContent = transaction.txid;
    status.className = "tx-status";
    status.textContent = transaction.status || "unknown";
    heading.className = "transaction-heading";
    heading.append(txid, status);

    meta.className = "transaction-meta";
    meta.textContent = `wallet: ${transaction.wallet || "--"} · confirmações: ${
      transaction.confirmations || 0
    } · idade: ${transaction.age_seconds || 0}s`;

    message.className = "transaction-message";
    message.textContent = transaction.message || transaction.warning || "";

    item.append(heading, meta, message);

    if (transaction.warning) {
      const warning = document.createElement("p");
      warning.className = "transaction-warning";
      warning.textContent = transaction.warning;
      item.append(warning);
    }

    ids.transactionList.append(item);
  }
}

async function refreshWallets() {
  setStatus("Atualizando...");

  try {
    const data = await fetchJSON("/wallets");
    updateWallets(data);
    if (data.selected_wallet) {
      await refreshWalletStatus();
    } else {
      updateWalletStatus(null);
    }
    await refreshTransactions();
    setStatus(`Atualizado ${new Date().toLocaleTimeString("pt-BR")}`);
  } catch (error) {
    setStatus("Erro ao atualizar", true);
    setWalletState(error.message.trim(), "error");
  }
}

async function selectWallet(wallet) {
  setStatus("Selecionando...");
  setWalletState(`Selecionando ${wallet}...`);

  const data = await fetchJSON("/wallet/select", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ wallet }),
  });

  setText(ids.selectedWallet, data.selected_wallet || "--");
  ids.selectedWallet.title = data.selected_wallet || "";
  updateWalletInfo(data.wallet_info || {});
  await refreshWalletStatus();
  await refreshTransactions();
  setWalletState(`Wallet selecionada: ${data.selected_wallet}`, "active");
  setStatus(`Atualizado ${new Date().toLocaleTimeString("pt-BR")}`);
}

async function refreshWalletStatus() {
  const status = await fetchJSON("/wallet/status");
  updateWalletStatus(status);
}

async function refreshTransactions() {
  const data = await fetchJSON("/txs");
  renderTransactions(data.transactions || []);
}

ids.walletForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const wallet = ids.walletSelect.value;
  if (!wallet) {
    return;
  }

  try {
    await selectWallet(wallet);
    await refreshWallets();
  } catch (error) {
    setStatus("Erro ao selecionar", true);
    setWalletState(error.message.trim(), "error");
  }
});

ids.walletSelect.addEventListener("change", async () => {
  const wallet = ids.walletSelect.value;
  if (!wallet) {
    return;
  }

  try {
    await selectWallet(wallet);
    await refreshWallets();
  } catch (error) {
    setStatus("Erro ao selecionar", true);
    setWalletState(error.message.trim(), "error");
  }
});

ids.sendForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const address = ids.sendAddress.value.trim();
  const amount = Number(ids.sendAmount.value);
  if (!address || !amount) {
    return;
  }

  ids.sendForm.querySelector("button").disabled = true;
  setSendState("Criando, assinando e transmitindo...", "active");

  try {
    const data = await fetchJSON("/tx/send", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ address, amount }),
    });

    ids.sendAddress.value = "";
    ids.sendAmount.value = "";
    setSendState(`Transação enviada: ${data.txid}`, "active");
    await refreshWalletStatus();
    await refreshTransactions();
  } catch (error) {
    setSendState(error.message.trim(), "error");
  } finally {
    ids.sendForm.querySelector("button").disabled = false;
  }
});

refreshWallets();
setInterval(refreshWallets, 10_000);
