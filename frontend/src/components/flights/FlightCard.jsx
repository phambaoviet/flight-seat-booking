import './FlightCard.css';
function FlightCard({ flight }) {
  const formatDuration = (duration) => {
    const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?/);

    if (!match) return duration;

    const hours = match[1] || 0;
    const minutes = match[2] || 0;

    return `${hours}h ${minutes}m`;
  };

  return (
    <article className="flight-card">
      <div className="flight-card-header">
        <div>
          <h3>
            {flight.slices?.[0]?.segments?.[0]?.marketing_carrier?.name ||
              'Airline'}
          </h3>
          <span>Flight offer</span>
        </div>

        <div className="flight-price">
          <strong>
            {flight.total_amount} {flight.total_currency}
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
