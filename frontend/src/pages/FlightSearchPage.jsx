import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';

import { searchFlights } from '../services/flightService';
import FlightCard from '../components/flights/FlightCard';
import './FlightSearchPage.css';

function FlightSearchPage() {
  const location = useLocation();

  const searchParams = location.state?.searchParams;

  const [flights, setFlights] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!searchParams) return;

    const fetchFlights = async () => {
      setLoading(true);
      setError(null);
      setFlights([]);

      try {
        const data = await searchFlights(searchParams);
        setFlights(data.flight_offers || []);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchFlights();
  }, [searchParams]);

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

      <div className="flight-results-content">
        {loading && (
          <div className="flight-loading">
            Searching for available flights...
          </div>
        )}

        {error && <p className="error-message">{error}</p>}

        {!loading && !error && flights.length === 0 && (
          <div className="flight-empty">No flights found for your search.</div>
        )}

        {!loading && flights.length > 0 && (
          <>
            <p className="flight-results-count">
              Found {flights.length} flight offers
            </p>

            <div className="flight-results-list">
              {flights.slice(0, 20).map((flight) => (
                <FlightCard key={flight.id} flight={flight} />
              ))}
            </div>
          </>
        )}
      </div>
    </main>
  );
}

export default FlightSearchPage;
