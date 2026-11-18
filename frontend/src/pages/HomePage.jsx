import Hero from '../components/home/Hero';
import PopularDestinations from '../components/home/PopularDestinations';
import WhyChooseUs from '../components/home/WhyChooseUs';
import FlightSearchSection from '../components/flights/FlightSearchSection';

function HomePage() {
  return (
    <>
      <Hero />
      <FlightSearchSection />
      <PopularDestinations />
      <WhyChooseUs />
    </>
  );
}

export default HomePage;
