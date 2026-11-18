import './PopularDestinations.css';

const destinations = [
  {
    code: 'HAN',
    city: 'Hanoi',
    country: 'Vietnam',
    image: '/images/hanoi.jpg',
  },
  {
    code: 'DAD',
    city: 'Da Nang',
    country: 'Vietnam',
    image: '/images/danang.jpg',
  },
  {
    code: 'SIN',
    city: 'Singapore',
    country: 'Singapore',
    image: '/images/singapore.jpg',
  },
];

function PopularDestinations() {
  return (
    <section className="destinations-section">
      <div className="section-heading">
        <span>EXPLORE THE WORLD</span>
        <h2>Popular Destinations</h2>
        <p>Discover amazing places for your next adventure.</p>
      </div>

      <div className="destinations-grid">
        {destinations.map((destination) => (
          <article className="destination-card" key={destination.code}>
            <img
              src={destination.image}
              alt={`${destination.city}, ${destination.country}`}
            />

            <div className="destination-info">
              <h3>{destination.city}</h3>
              <p>{destination.country}</p>
            </div>
          </article>
        ))}
      </div>
    </section>
  );
}

export default PopularDestinations;
