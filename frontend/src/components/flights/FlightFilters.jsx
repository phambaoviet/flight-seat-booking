import './FlightFilters.css';
function FlightFilters({
  airlines,
  selectedAirline,
  onAirlineChange,
  maxPrice,
  onMaxPriceChange,
  sortBy,
  onSortChange,
}) {
  return (
    <aside className="flight-filters">
      <h3>Filters</h3>

      <div className="filter-group">
        <label htmlFor="airline-filter">Airline</label>

        <select
          id="airline-filter"
          value={selectedAirline}
          onChange={(e) => onAirlineChange(e.target.value)}
        >
          <option value="all">All airlines</option>

          {airlines.map((airline) => (
            <option key={airline} value={airline}>
              {airline}
            </option>
          ))}
        </select>
      </div>

      <div className="filter-group">
        <label htmlFor="max-price">Maximum price</label>

        <input
          id="max-price"
          type="number"
          min="0"
          placeholder="No limit"
          value={maxPrice}
          onChange={(e) => onMaxPriceChange(e.target.value)}
        />
      </div>

      <div className="filter-group">
        <label htmlFor="sort-by">Sort by</label>

        <select
          id="sort-by"
          value={sortBy}
          onChange={(e) => onSortChange(e.target.value)}
        >
          <option value="recommended">Recommended</option>
          <option value="price-asc">Price: Low to High</option>
          <option value="price-desc">Price: High to Low</option>
          <option value="duration-asc">Shortest Duration</option>
        </select>
      </div>
    </aside>
  );
}

export default FlightFilters;
