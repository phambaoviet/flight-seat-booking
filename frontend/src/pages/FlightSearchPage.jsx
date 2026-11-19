import { useEffect, useMemo, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { searchFlights } from '../services/flightService';
import { convertToVND } from '../services/currencyService';
import FlightCard from '../components/flights/FlightCard';
import './FlightSearchPage.css';

const PRICE_RANGES = [
  { label: 'Dưới 2 triệu', min: 0, max: 2000000 },
  { label: '2 – 4 triệu', min: 2000000, max: 4000000 },
  { label: '4 – 6 triệu', min: 4000000, max: 6000000 },
  { label: '6 – 10 triệu', min: 6000000, max: 10000000 },
  { label: 'Trên 10 triệu', min: 10000000, max: Infinity },
];

function FlightSearchPage() {
  const location = useLocation();
  const searchParams = location.state?.searchParams;

  const [flights, setFlights] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currencyWarning, setCurrencyWarning] = useState('');
  const [selectedRanges, setSelectedRanges] = useState([]);
  const [sortOrder, setSortOrder] = useState('cheapest');

  useEffect(() => {
    if (!searchParams) return;

    let cancelled = false;

    const fetchFlights = async () => {
      setLoading(true);
      setError(null);
      setCurrencyWarning('');
      setFlights([]);

      try {
        const data = await searchFlights(searchParams);
        const offers = data.flight_offers || [];

        const convertedOffers = await Promise.all(
          offers.map(async (flight) => {
            try {
              const priceVND = await convertToVND(
                flight.total_amount,
                flight.total_currency
              );

              return { ...flight, priceVND };
            } catch {
              return { ...flight, priceVND: null };
            }
          })
        );

        if (cancelled) return;

        setFlights(convertedOffers);

        if (convertedOffers.some((flight) => flight.priceVND === null)) {
          setCurrencyWarning(
            'Một số vé chưa thể quy đổi sang VND. Giá gốc vẫn được giữ lại.'
          );
        }
      } catch (err) {
        if (!cancelled) setError(err.message);
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    fetchFlights();

    return () => {
      cancelled = true;
    };
  }, [searchParams]);

  const filteredFlights = useMemo(() => {
    let result = [...flights];

    if (selectedRanges.length > 0) {
      result = result.filter((flight) =>
        selectedRanges.some((index) => {
          const range = PRICE_RANGES[index];

          return (
            flight.priceVND !== null &&
            flight.priceVND >= range.min &&
            flight.priceVND < range.max
          );
        })
      );
    }

    result.sort((a, b) => {
      if (a.priceVND === null) return 1;
      if (b.priceVND === null) return -1;

      return sortOrder === 'cheapest'
        ? a.priceVND - b.priceVND
        : b.priceVND - a.priceVND;
    });

    return result;
  }, [flights, selectedRanges, sortOrder]);

  const handleRangeChange = (index) => {
    setSelectedRanges((previous) =>
      previous.includes(index)
        ? previous.filter((item) => item !== index)
        : [...previous, index]
    );
  };

  if (!searchParams) {
    return (
      <main className="flight-results-page">
        <h1>Flight Search</h1>
        <p>Please search for a flight from the home page.</p>
      </main>
    );
  }

  return (
    <main className="flight-results-page">
      <div className="flight-results-heading">
        <span>YOUR JOURNEY</span>
        <h1>Available Flights</h1>
        <p>
          {searchParams.origin.toUpperCase()} →{' '}
          {searchParams.destination.toUpperCase()}
        </p>
      </div>

      {loading && <p>Searching for available flights...</p>}
      {error && <p className="error-message">{error}</p>}

      {currencyWarning && <p className="currency-warning">{currencyWarning}</p>}

      {!loading && !error && (
        <div className="flight-results-layout">
          <aside className="price-filter">
            <h3>Filter by price</h3>

            {PRICE_RANGES.map((range, index) => (
              <label key={range.label} className="price-filter-option">
                <input
                  type="checkbox"
                  checked={selectedRanges.includes(index)}
                  onChange={() => handleRangeChange(index)}
                />
                <span>{range.label} ₫</span>
              </label>
            ))}

            <button
              type="button"
              className="clear-filter"
              onClick={() => setSelectedRanges([])}
            >
              Clear filters
            </button>
          </aside>

          <section className="flight-results-content">
            <div className="flight-results-toolbar">
              <p>Found {filteredFlights.length} flight offers</p>

              <select
                value={sortOrder}
                onChange={(e) => setSortOrder(e.target.value)}
              >
                <option value="cheapest">Price: Lowest first</option>
                <option value="expensive">Price: Highest first</option>
              </select>
            </div>

            {filteredFlights.length === 0 ? (
              <p>No flights found for this price range.</p>
            ) : (
              <div className="flight-results-list">
                {filteredFlights.map((flight) => (
                  <FlightCard key={flight.id} flight={flight} />
                ))}
              </div>
            )}
          </section>
        </div>
      )}
    </main>
  );
}

export default FlightSearchPage;
