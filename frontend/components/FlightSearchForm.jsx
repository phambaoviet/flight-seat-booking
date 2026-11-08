import React, { useState } from 'react';

const initialFormState = {
  tripType: 'round-trip',
  origin: '',
  destination: '',
  departureDate: '',
  returnDate: '',
  passengers: 1,
  travelClass: 'economy',
};

function FlightSearchForm({ onSearch }) {
  const [formData, setFormData] = useState(initialFormState);
  const [errors, setErrors] = useState({});
  const handleTripTypeChange = (e) => {
    const { value } = e.target;
    setFormData((prev) => ({
      ...prev,
      tripType: value,
      returnDate: value === 'one-way' ? '' : prev.returnDate,
    }));
  };

  const handleChange = (e) => {
    const { name, value } = e.target;

    setFormData((prev) => ({
      ...prev,
      [name]: name === 'passengers' ? Number(value) : value,
    }));
  };

  const handleSubmit = (e) => {
    // 1. Ngăn form reload trang.
    e.preventDefault();
    // 2. Gọi validateForm() và lưu kết quả vào biến newErrors.
    const newErrors = validateForm();
    // 3. Cập nhật errors state bằng setErrors().
    setErrors(newErrors);
    // 4. Nếu newErrors có lỗi → dừng hàm.
    if (Object.keys(newErrors).length > 0) {
      return;
    }

    // 5. Nếu không có lỗi → gọi onSearch(formData).
    onSearch(formData);
  };

  const validateForm = () => {
    const newErrors = {};

    // Nếu origin trống thì thêm lỗi vào newErrors.origin
    if (!formData.origin) {
      newErrors.origin = 'Please enter departure location';
    }
    // Nếu destination trống thì thêm lỗi vào newErrors.destination
    if (!formData.destination) {
      newErrors.destination = 'Please enter destination';
    }
    // Nếu departureDate trống thì thêm lỗi vào newErrors.departureDate
    if (!formData.departureDate) {
      newErrors.departureDate = 'Please select a departure date';
    }
    // Tiếp tục kiểm tra returnDate, passengers và thứ tự ngày...
    if (formData.tripType === 'round-trip' && !formData.returnDate) {
      newErrors.returnDate =
        'Please select a return date for round-trip flights.';
    }
    if (!Number.isInteger(formData.passengers) || formData.passengers < 1) {
      newErrors.passengers = 'Number of passengers must be at least 1.';
    }
    if (
      formData.tripType === 'round-trip' &&
      formData.returnDate < formData.departureDate && 
      formData.departureDate &&
      formData.returnDate
    ) {
      newErrors.returnDate =
        'Return date cannot be earlier than departure date.';
    }
    return newErrors;
  };

  return (
    <form onSubmit={handleSubmit}>
      <div>
        <label>
          Trip Type:
          <select
            name="tripType"
            value={formData.tripType}
            onChange={handleTripTypeChange}
          >
            <option value="round-trip">Round Trip</option>
            <option value="one-way">One Way</option>
          </select>
        </label>
      </div>
      <div>
        <label>
          Origin:
          <input
            type="text"
            name="origin"
            value={formData.origin}
            onChange={handleChange}
            required
          />
        </label>
        {errors.origin && <p className="error-message">{errors.origin}</p>}
      </div>
      <div>
        <label>
          Destination:
          <input
            type="text"
            name="destination"
            value={formData.destination}
            onChange={handleChange}
            required
          />
        </label>
        {errors.destination && (
          <p className="error-message">{errors.destination}</p>
        )}
      </div>
      <div>
        <label>
          Departure Date:
          <input
            type="date"
            name="departureDate"
            value={formData.departureDate}
            onChange={handleChange}
            required
          />
        </label>
        {errors.departureDate && (
          <p className="error-message">{errors.departureDate}</p>
        )}
      </div>
      {formData.tripType === 'round-trip' && (
        <div>
          <label>
            Return Date:
            <input
              type="date"
              name="returnDate"
              value={formData.returnDate}
              onChange={handleChange}
              min={formData.departureDate}
              required
            />
          </label>
          {errors.returnDate && (
            <p className="error-message">{errors.returnDate}</p>
          )}
        </div>
      )}
      <div>
        <label>
          Passengers:
          <input
            type="number"
            name="passengers"
            value={formData.passengers}
            onChange={handleChange}
            min="1"
            required
          />
        </label>
        {errors.passengers && (
          <p className="error-message">{errors.passengers}</p>
        )}
      </div>
      <div>
        <label>
          Travel Class:
          <select
            name="travelClass"
            value={formData.travelClass}
            onChange={handleChange}
          >
            <option value="economy">Economy</option>
            <option value="business">Business</option>
            <option value="first">First Class</option>
          </select>
        </label>
      </div>
      <button type="submit">Search Flights</button>
    </form>
  );
}

export default FlightSearchForm;
