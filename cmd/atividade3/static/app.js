const ids = {
  refreshStatus: document.querySelector("#refresh-status"),
  selectedWallet: document.querySelector("#selected-wallet"),
  walletForm: document.querySelector("#wallet-form"),
  walletSelect: document.querySelector("#wallet-select"),
  walletState: document.querySelector("#wallet-state"),
  walletBalance: document.querySelector("#wallet-balance"),
  walletName: document.querySelector("#wallet-name"),
  walletTxCount: document.querySelector("#wallet-txcount"),
  availableWallets: document.querySelector("#available-wallets"),
  loadedWallets: document.querySelector("#loaded-wallets"),
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

async function refreshWallets() {
  setStatus("Atualizando...");

  try {
    const data = await fetchJSON("/wallets");
    updateWallets(data);
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
  setWalletState(`Wallet selecionada: ${data.selected_wallet}`, "active");
  setStatus(`Atualizado ${new Date().toLocaleTimeString("pt-BR")}`);
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

refreshWallets();
setInterval(refreshWallets, 10_000);
