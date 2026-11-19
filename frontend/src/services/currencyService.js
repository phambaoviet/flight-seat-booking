const rateCache = new Map();

export async function getExchangeRate(from, to = 'VND') {
  const source = from.toUpperCase();
  const target = to.toUpperCase();

  if (source === target) return 1;

  const cacheKey = `${source}_${target}`;

  if (rateCache.has(cacheKey)) {
    return rateCache.get(cacheKey);
  }

  const response = await fetch(
    `https://api.frankfurter.dev/v2/rate/${source.toLowerCase()}/${target.toLowerCase()}`
  );

  if (!response.ok) {
    throw new Error(`Unable to get exchange rate: ${source}/${target}`);
  }

  const data = await response.json();

  if (typeof data.rate !== 'number') {
    throw new Error('Invalid exchange rate data');
  }

  rateCache.set(cacheKey, data.rate);

  return data.rate;
}

export async function convertToVND(amount, currency) {
  const rate = await getExchangeRate(currency, 'VND');

  return Number(amount) * rate;
}

export function formatVND(amount) {
  return (
    new Intl.NumberFormat('vi-VN', {
      maximumFractionDigits: 0,
    }).format(amount) + ' ₫'
  );
}
