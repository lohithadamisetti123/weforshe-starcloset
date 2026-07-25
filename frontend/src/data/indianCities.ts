// Comprehensive city × state dataset for the onboarding location selector.
// Display format everywhere is "City, State"; search matches either part.
// This is UI data only — the profile stores the plain city name, which is what
// the backend's location signal recognises (unknown cities are simply neutral).

export interface IndianCity {
  city: string
  state: string
}

export const INDIAN_CITIES: IndianCity[] = [
  // Andhra Pradesh
  { city: 'Visakhapatnam', state: 'Andhra Pradesh' },
  { city: 'Vijayawada', state: 'Andhra Pradesh' },
  { city: 'Guntur', state: 'Andhra Pradesh' },
  { city: 'Nellore', state: 'Andhra Pradesh' },
  { city: 'Tirupati', state: 'Andhra Pradesh' },
  { city: 'Rajahmundry', state: 'Andhra Pradesh' },
  { city: 'Kakinada', state: 'Andhra Pradesh' },
  { city: 'Kurnool', state: 'Andhra Pradesh' },
  { city: 'Anantapur', state: 'Andhra Pradesh' },
  // Arunachal Pradesh
  { city: 'Itanagar', state: 'Arunachal Pradesh' },
  // Assam
  { city: 'Guwahati', state: 'Assam' },
  { city: 'Silchar', state: 'Assam' },
  { city: 'Dibrugarh', state: 'Assam' },
  { city: 'Jorhat', state: 'Assam' },
  { city: 'Tezpur', state: 'Assam' },
  // Bihar
  { city: 'Patna', state: 'Bihar' },
  { city: 'Gaya', state: 'Bihar' },
  { city: 'Bhagalpur', state: 'Bihar' },
  { city: 'Muzaffarpur', state: 'Bihar' },
  { city: 'Darbhanga', state: 'Bihar' },
  // Chhattisgarh
  { city: 'Raipur', state: 'Chhattisgarh' },
  { city: 'Bhilai', state: 'Chhattisgarh' },
  { city: 'Bilaspur', state: 'Chhattisgarh' },
  { city: 'Korba', state: 'Chhattisgarh' },
  // Goa
  { city: 'Panaji', state: 'Goa' },
  { city: 'Margao', state: 'Goa' },
  { city: 'Vasco da Gama', state: 'Goa' },
  { city: 'Mapusa', state: 'Goa' },
  // Gujarat
  { city: 'Ahmedabad', state: 'Gujarat' },
  { city: 'Surat', state: 'Gujarat' },
  { city: 'Vadodara', state: 'Gujarat' },
  { city: 'Rajkot', state: 'Gujarat' },
  { city: 'Bhavnagar', state: 'Gujarat' },
  { city: 'Jamnagar', state: 'Gujarat' },
  { city: 'Gandhinagar', state: 'Gujarat' },
  { city: 'Junagadh', state: 'Gujarat' },
  // Haryana
  { city: 'Gurugram', state: 'Haryana' },
  { city: 'Faridabad', state: 'Haryana' },
  { city: 'Panipat', state: 'Haryana' },
  { city: 'Ambala', state: 'Haryana' },
  { city: 'Hisar', state: 'Haryana' },
  { city: 'Karnal', state: 'Haryana' },
  // Himachal Pradesh
  { city: 'Shimla', state: 'Himachal Pradesh' },
  { city: 'Dharamshala', state: 'Himachal Pradesh' },
  { city: 'Solan', state: 'Himachal Pradesh' },
  { city: 'Mandi', state: 'Himachal Pradesh' },
  // Jharkhand
  { city: 'Ranchi', state: 'Jharkhand' },
  { city: 'Jamshedpur', state: 'Jharkhand' },
  { city: 'Dhanbad', state: 'Jharkhand' },
  { city: 'Bokaro Steel City', state: 'Jharkhand' },
  // Karnataka
  { city: 'Bengaluru', state: 'Karnataka' },
  { city: 'Mysuru', state: 'Karnataka' },
  { city: 'Mangaluru', state: 'Karnataka' },
  { city: 'Hubballi', state: 'Karnataka' },
  { city: 'Belagavi', state: 'Karnataka' },
  { city: 'Davanagere', state: 'Karnataka' },
  { city: 'Shivamogga', state: 'Karnataka' },
  { city: 'Udupi', state: 'Karnataka' },
  // Kerala
  { city: 'Thiruvananthapuram', state: 'Kerala' },
  { city: 'Kochi', state: 'Kerala' },
  { city: 'Kozhikode', state: 'Kerala' },
  { city: 'Thrissur', state: 'Kerala' },
  { city: 'Kollam', state: 'Kerala' },
  { city: 'Kannur', state: 'Kerala' },
  { city: 'Alappuzha', state: 'Kerala' },
  // Madhya Pradesh
  { city: 'Bhopal', state: 'Madhya Pradesh' },
  { city: 'Indore', state: 'Madhya Pradesh' },
  { city: 'Gwalior', state: 'Madhya Pradesh' },
  { city: 'Jabalpur', state: 'Madhya Pradesh' },
  { city: 'Ujjain', state: 'Madhya Pradesh' },
  { city: 'Sagar', state: 'Madhya Pradesh' },
  { city: 'Rewa', state: 'Madhya Pradesh' },
  // Maharashtra
  { city: 'Mumbai', state: 'Maharashtra' },
  { city: 'Pune', state: 'Maharashtra' },
  { city: 'Nagpur', state: 'Maharashtra' },
  { city: 'Nashik', state: 'Maharashtra' },
  { city: 'Thane', state: 'Maharashtra' },
  { city: 'Navi Mumbai', state: 'Maharashtra' },
  { city: 'Chhatrapati Sambhajinagar', state: 'Maharashtra' },
  { city: 'Solapur', state: 'Maharashtra' },
  { city: 'Amravati', state: 'Maharashtra' },
  { city: 'Kolhapur', state: 'Maharashtra' },
  // Manipur
  { city: 'Imphal', state: 'Manipur' },
  // Meghalaya
  { city: 'Shillong', state: 'Meghalaya' },
  // Mizoram
  { city: 'Aizawl', state: 'Mizoram' },
  // Nagaland
  { city: 'Kohima', state: 'Nagaland' },
  { city: 'Dimapur', state: 'Nagaland' },
  // Odisha
  { city: 'Bhubaneswar', state: 'Odisha' },
  { city: 'Cuttack', state: 'Odisha' },
  { city: 'Rourkela', state: 'Odisha' },
  { city: 'Sambalpur', state: 'Odisha' },
  { city: 'Puri', state: 'Odisha' },
  // Punjab
  { city: 'Ludhiana', state: 'Punjab' },
  { city: 'Amritsar', state: 'Punjab' },
  { city: 'Jalandhar', state: 'Punjab' },
  { city: 'Patiala', state: 'Punjab' },
  { city: 'Bathinda', state: 'Punjab' },
  { city: 'Mohali', state: 'Punjab' },
  // Rajasthan
  { city: 'Jaipur', state: 'Rajasthan' },
  { city: 'Jodhpur', state: 'Rajasthan' },
  { city: 'Udaipur', state: 'Rajasthan' },
  { city: 'Kota', state: 'Rajasthan' },
  { city: 'Ajmer', state: 'Rajasthan' },
  { city: 'Bikaner', state: 'Rajasthan' },
  { city: 'Alwar', state: 'Rajasthan' },
  // Sikkim
  { city: 'Gangtok', state: 'Sikkim' },
  // Tamil Nadu
  { city: 'Chennai', state: 'Tamil Nadu' },
  { city: 'Coimbatore', state: 'Tamil Nadu' },
  { city: 'Madurai', state: 'Tamil Nadu' },
  { city: 'Tiruchirappalli', state: 'Tamil Nadu' },
  { city: 'Salem', state: 'Tamil Nadu' },
  { city: 'Tirunelveli', state: 'Tamil Nadu' },
  { city: 'Erode', state: 'Tamil Nadu' },
  { city: 'Vellore', state: 'Tamil Nadu' },
  { city: 'Thoothukudi', state: 'Tamil Nadu' },
  // Telangana
  { city: 'Hyderabad', state: 'Telangana' },
  { city: 'Warangal', state: 'Telangana' },
  { city: 'Nizamabad', state: 'Telangana' },
  { city: 'Karimnagar', state: 'Telangana' },
  { city: 'Khammam', state: 'Telangana' },
  // Tripura
  { city: 'Agartala', state: 'Tripura' },
  // Uttar Pradesh
  { city: 'Lucknow', state: 'Uttar Pradesh' },
  { city: 'Kanpur', state: 'Uttar Pradesh' },
  { city: 'Varanasi', state: 'Uttar Pradesh' },
  { city: 'Agra', state: 'Uttar Pradesh' },
  { city: 'Prayagraj', state: 'Uttar Pradesh' },
  { city: 'Meerut', state: 'Uttar Pradesh' },
  { city: 'Ghaziabad', state: 'Uttar Pradesh' },
  { city: 'Noida', state: 'Uttar Pradesh' },
  { city: 'Bareilly', state: 'Uttar Pradesh' },
  { city: 'Aligarh', state: 'Uttar Pradesh' },
  { city: 'Moradabad', state: 'Uttar Pradesh' },
  { city: 'Gorakhpur', state: 'Uttar Pradesh' },
  { city: 'Jhansi', state: 'Uttar Pradesh' },
  { city: 'Mathura', state: 'Uttar Pradesh' },
  // Uttarakhand
  { city: 'Dehradun', state: 'Uttarakhand' },
  { city: 'Haridwar', state: 'Uttarakhand' },
  { city: 'Rishikesh', state: 'Uttarakhand' },
  { city: 'Nainital', state: 'Uttarakhand' },
  { city: 'Haldwani', state: 'Uttarakhand' },
  // West Bengal
  { city: 'Kolkata', state: 'West Bengal' },
  { city: 'Howrah', state: 'West Bengal' },
  { city: 'Durgapur', state: 'West Bengal' },
  { city: 'Asansol', state: 'West Bengal' },
  { city: 'Siliguri', state: 'West Bengal' },
  { city: 'Darjeeling', state: 'West Bengal' },
  // Union Territories
  { city: 'New Delhi', state: 'Delhi' },
  { city: 'Delhi', state: 'Delhi' },
  { city: 'Chandigarh', state: 'Chandigarh' },
  { city: 'Puducherry', state: 'Puducherry' },
  { city: 'Srinagar', state: 'Jammu & Kashmir' },
  { city: 'Jammu', state: 'Jammu & Kashmir' },
  { city: 'Leh', state: 'Ladakh' },
  { city: 'Port Blair', state: 'Andaman & Nicobar Islands' },
  { city: 'Daman', state: 'Dadra & Nagar Haveli and Daman & Diu' },
  { city: 'Silvassa', state: 'Dadra & Nagar Haveli and Daman & Diu' },
  { city: 'Kavaratti', state: 'Lakshadweep' },
]

export const cityDisplay = (c: IndianCity) => `${c.city}, ${c.state}`
