import './Hero.css';

function Hero() {
  return (
    <section className="hero-section">
      <div className="hero-content">
        <span className="hero-label">✦ YOUR JOURNEY, OUR PRIORITY</span>

        <h1>
          Your Journey
          <br />
          Begins <span>Here.</span>
        </h1>

        <p className="hero-description">
          Discover new destinations and enjoy a seamless flight booking
          experience with SkyBook.
        </p>

        <a href="#flight-search" className="hero-button">
          Explore Flights
          <span>→</span>
        </a>

        <div className="hero-stats">
          <div>
            <strong>100+</strong>
            <span>Destinations</span>
          </div>

          <div>
            <strong>24/7</strong>
            <span>Customer Support</span>
          </div>

          <div>
            <strong>Easy</strong>
            <span>Booking Experience</span>
          </div>
        </div>
      </div>

      <div className="hero-visual">
        <img
          src="https://images.unsplash.com/photo-1436491865332-7a61a109cc05?w=1200"
          alt="Airplane flying above the clouds"
          className="hero-image"
        />

        <div className="hero-image-overlay"></div>

        <div className="hero-caption">
          <span>✈</span>
          <div>
            <h3>Explore the World</h3>
            <p>Your next journey starts here.</p>
          </div>
        </div>

        <div className="hero-floating-card">
          <span className="floating-icon">✦</span>
          <div>
            <strong>Travel with confidence</strong>
            <p>Your journey, made simple.</p>
          </div>
        </div>
      </div>
    </section>
  );
}

export default Hero;
