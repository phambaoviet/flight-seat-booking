import { useNavigate } from 'react-router-dom';
import FlightSearchForm from './FlightSearchForm';
import './FlightSearchSection.css';

function FlightSearchSection() {
  const navigate = useNavigate();

  const handleSearch = (formData) => {
    navigate('/flights/search', {
      state: {
        searchParams: formData,
      },
    });
  };

  return (
    <section id="flight-search" className="flight-search-section">
      <div className="search-heading">
        <span>BOOK YOUR NEXT FLIGHT</span>
        <h2>Find your perfect flight</h2>
        <p>Search destinations and discover flights for your next journey.</p>
      </div>

      <div className="search-card">
        <FlightSearchForm onSearch={handleSearch} />
      </div>
    </section>
  );
}

export default FlightSearchSection;
