import './WhyChooseUs.css';

const features = [
  {
    icon: '✈',
    title: 'Easy Flight Booking',
    description:
      'Search and book flights with a simple and user-friendly experience.',
  },
  {
    icon: '🔒',
    title: 'Secure Reservations',
    description: 'Your booking information is handled with security and care.',
  },
  {
    icon: '🌍',
    title: 'Explore More Destinations',
    description:
      'Discover destinations and find flights for your next adventure.',
  },
  {
    icon: '💬',
    title: 'Customer Support',
    description: 'Get assistance whenever you need help with your journey.',
  },
];

function WhyChooseUs() {
  return (
    <section className="why-section">
      <div className="section-heading">
        <span>WHY SKYBOOK</span>
        <h2>Travel with Confidence</h2>
        <p>Everything you need for a smoother booking experience.</p>
      </div>

      <div className="features-grid">
        {features.map((feature) => (
          <article className="feature-card" key={feature.title}>
            <div className="feature-icon">{feature.icon}</div>
            <h3>{feature.title}</h3>
            <p>{feature.description}</p>
          </article>
        ))}
      </div>
    </section>
  );
}

export default WhyChooseUs;
