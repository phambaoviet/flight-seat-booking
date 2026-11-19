const API_BASE_URL = '/api/v1';

export async function searchFlights(formData) {
  const payload = {
    origin: formData.origin.trim().toUpperCase(),
    destination: formData.destination.trim().toUpperCase(),
    departure: formData.departureDate,
    return: formData.tripType === 'round-trip' ? formData.returnDate : '',
    adults: formData.adults,
    children: formData.children,
    infants: formData.infants,
    cabin_class: formData.travelClass,
    trip_type: formData.tripType,
  };

  const response = await fetch(`${API_BASE_URL}/flights/search`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.message || data.error || 'Unable to search flights.');
  }

  return data;
}
