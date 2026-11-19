import { useEffect, useState } from 'react';
import './FlightCard.css';
import { convertToVND, formatVND } from '../../services/currencyService';

function FlightCard({ flight }) {
  const [priceVND, setPriceVND] = useState(null);
  const [priceError, setPriceError] = useState(false);

  // Convert original price to VND
  useEffect(() => {
    let cancelled = false;

    const fetchPrice = async () => {
      try {
        const convertedPrice = await convertToVND(
          flight.total_amount,
          flight.total_currency
        );

        if (!cancelled) {
          setPriceVND(convertedPrice);
          setPriceError(false);
        }
      } catch (error) {
        if (!cancelled) {
          setPriceError(true);
        }
      }
    };

    fetchPrice();

    return () => {
      cancelled = true;
    };
  }, [flight.total_amount, flight.total_currency]);

  // Convert ISO 8601 duration to readable format
  const formatDuration = (duration) => {
    const match = duration?.match(/PT(?:(\d+)H)?(?:(\d+)M)?/);

    if (!match) return duration || 'N/A';

    const hours = match[1] || 0;
    const minutes = match[2] || 0;

    return `${hours}h ${minutes}m`;
  };

  const getAirlineName = (slice) => {
    return slice?.segments?.[0]?.marketing_carrier?.name || 'Airline';
  };

  return (
    <article className="flight-card">
      <div className="flight-card-header">
        <div>
          <h3>{getAirlineName(flight.slices?.[0])}</h3>
          <span>Flight offer</span>
        </div>

        <div className="flight-price">
          <strong className="flight-card-price">
            {priceVND !== null
              ? formatVND(priceVND)
              : priceError
              ? 'Unable to convert'
              : 'Converting...'}
          </strong>
        </div>
      </div>

      {flight.slices?.map((slice, index) => (
        <div className="flight-slice" key={index}>
          <h4>{index === 0 ? 'Outbound flight' : 'Return flight'}</h4>

          <div className="flight-route">
            <div>
              <strong>{slice.origin?.iata}</strong>
              <p>{slice.origin?.name}</p>
            </div>

            <div className="flight-route-details">
              {slice.segments?.map((segment) => (
                <div key={segment.id}>
                  <span>{segment.marketing_carrier?.name}</span>
                  <span>{formatDuration(segment.duration)}</span>
                </div>
              ))}
            </div>

            <div>
              <strong>{slice.destination?.iata}</strong>
              <p>{slice.destination?.name}</p>
            </div>
          </div>
        </div>
      ))}

      <button type="button" className="flight-select-button">
        Select Flight
      </button>
    </article>
  );
}

export default FlightCard;
