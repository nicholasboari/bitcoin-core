const ids = {
  refreshStatus: document.querySelector("#refresh-status"),
  txCount: document.querySelector("#tx-count"),
  avgFeeRate: document.querySelector("#avg-fee-rate"),
  totalVSize: document.querySelector("#total-vsize"),
  feeLow: document.querySelector("#fee-low"),
  feeMedium: document.querySelector("#fee-medium"),
  feeHigh: document.querySelector("#fee-high"),
  feeLowMeter: document.querySelector("#fee-low-meter"),
  feeMediumMeter: document.querySelector("#fee-medium-meter"),
  feeHighMeter: document.querySelector("#fee-high-meter"),
  syncLag: document.querySelector("#sync-lag"),
  blocks: document.querySelector("#blocks"),
  headers: document.querySelector("#headers"),
  syncState: document.querySelector("#sync-state"),
};

const numberFormat = new Intl.NumberFormat("pt-BR");
const feeFormat = new Intl.NumberFormat("pt-BR", {
  maximumFractionDigits: 2,
});

async function fetchJSON(url) {
  const response = await fetch(url, { headers: { Accept: "application/json" } });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `HTTP ${response.status}`);
  }

  return response.json();
}

function setText(element, value) {
  element.textContent = value;
}

function updateMempool(summary) {
  const distribution = summary.fee_distribution || {};
  const low = distribution.low || 0;
  const medium = distribution.medium || 0;
  const high = distribution.high || 0;
  const maxBucket = Math.max(low, medium, high, 1);

  setText(ids.txCount, numberFormat.format(summary.tx_count || 0));
  setText(ids.avgFeeRate, feeFormat.format(summary.avg_fee_rate || 0));
  setText(ids.totalVSize, numberFormat.format(summary.total_vsize || 0));
  setText(ids.feeLow, numberFormat.format(low));
  setText(ids.feeMedium, numberFormat.format(medium));
  setText(ids.feeHigh, numberFormat.format(high));

  ids.feeLowMeter.max = maxBucket;
  ids.feeMediumMeter.max = maxBucket;
  ids.feeHighMeter.max = maxBucket;
  ids.feeLowMeter.value = low;
  ids.feeMediumMeter.value = medium;
  ids.feeHighMeter.value = high;
}

function updateSync(info) {
  const lag = info.blocks_to_headers || 0;

  setText(ids.syncLag, numberFormat.format(lag));
  setText(ids.blocks, numberFormat.format(info.blocks || 0));
  setText(ids.headers, numberFormat.format(info.headers || 0));

  ids.syncState.classList.remove("synced", "lagging", "error");

  if (lag === 0) {
    ids.syncState.classList.add("synced");
    setText(ids.syncState, "Sincronizado");
  } else {
    ids.syncState.classList.add("lagging");
    setText(ids.syncState, `${numberFormat.format(lag)} bloco(s) de atraso`);
  }
}

async function refreshDashboard() {
  ids.refreshStatus.classList.remove("error");
  setText(ids.refreshStatus, "Atualizando...");

  try {
    const [mempool, blockchain] = await Promise.all([
      fetchJSON("/api/mempool/summary"),
      fetchJSON("/api/blockchain/info"),
    ]);

    updateMempool(mempool);
    updateSync(blockchain);
    setText(ids.refreshStatus, `Atualizado ${new Date().toLocaleTimeString("pt-BR")}`);
  } catch (error) {
    ids.refreshStatus.classList.add("error");
    ids.syncState.classList.add("error");
    setText(ids.refreshStatus, "Erro ao atualizar");
    setText(ids.syncState, error.message.trim());
  }
}

refreshDashboard();
setInterval(refreshDashboard, 10_000);
