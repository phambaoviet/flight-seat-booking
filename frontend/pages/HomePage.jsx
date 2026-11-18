import Hero from '../components/Hero';
import FlightSearchSection from '../components/FlightSearchSection';
import PopularDestinations from '../components/home/PopularDestinations';
import WhyChooseUs from '../components/home/WhyChooseUs';

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
