import React, { useState } from 'react';
import FlightSearchForm from '../components/FlightSearchForm';
function FlightSearchPage() {
  const [flights, setFlights] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const handleSearch = (formData) => {
    setLoading(true);
    setError(null);
    // Call the backend API to search for flights based on formData
    fetch('/api/flights/search', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(formData),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error('Failed to fetch flight results');
        }
        return response.json();
      })
      .then((data) => {
        setFlights(data);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  };

  return (
    <div>
      <h1>Flight Search</h1>
      {/* Render the FlightSearchForm component and pass the handleSearch function as a prop */}
      <FlightSearchForm onSearch={handleSearch} />
      {loading && <p>Loading flight results...</p>}
      {error && <p>Error: {error}</p>}
      {flights.length > 0 && (
        <div>
          <h2>Flight Results:</h2>
          <ul>
            {flights.map((flight) => (
              <li key={flight.id}>
                {flight.airline} - ${flight.price}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default FlightSearchPage;
