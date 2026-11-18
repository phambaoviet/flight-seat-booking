import { Link } from 'react-router-dom';
import './AboutPage.css';

const values = [
  {
    icon: '🎯',
    title: 'Simplicity',
    description: 'We make flight searching and booking simple and convenient.',
  },
  {
    icon: '🔒',
    title: 'Security',
    description: 'We prioritize the security and privacy of our customers.',
  },
  {
    icon: '🌍',
    title: 'Accessibility',
    description:
      'We help travelers discover destinations and plan their journeys.',
  },
];

function AboutPage() {
  return (
    <main className="about-page">
      <section className="about-hero">
        <span className="about-eyebrow">ABOUT SKYBOOK</span>
        <h1>
          Making Every Journey
          <span> Easier.</span>
        </h1>
        <p>
          SkyBook is a flight booking platform designed to make discovering
          destinations and planning your next journey a simpler experience.
        </p>
        <Link to="/" className="about-cta">
          Explore Flights →
        </Link>
      </section>

      <section className="about-mission">
        <div className="mission-content">
          <span className="about-eyebrow">WHO WE ARE</span>
          <h2>Travel should be simple.</h2>
          <p>
            We believe planning a trip should be an enjoyable experience.
            SkyBook brings flight discovery and booking into one convenient
            platform.
          </p>
        </div>

        <div className="mission-cards">
          <article className="mission-card">
            <span>🎯</span>
            <h3>Our Mission</h3>
            <p>
              To provide a convenient and user-friendly flight booking
              experience.
            </p>
          </article>

          <article className="mission-card">
            <span>🚀</span>
            <h3>Our Vision</h3>
            <p>To make travel planning more accessible through technology.</p>
          </article>
        </div>
      </section>

      <section className="about-values">
        <div className="section-heading">
          <span>WHAT MATTERS TO US</span>
          <h2>Our Core Values</h2>
          <p>The principles behind the SkyBook experience.</p>
        </div>

        <div className="values-grid">
          {values.map((value) => (
            <article className="value-card" key={value.title}>
              <div className="value-icon">{value.icon}</div>
              <h3>{value.title}</h3>
              <p>{value.description}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="about-cta-section">
        <h2>Ready for your next journey?</h2>
        <p>Discover your next destination with SkyBook.</p>
        <Link to="/" className="about-cta">
          Start Exploring →
        </Link>
      </section>
    </main>
  );
}

export default AboutPage;
