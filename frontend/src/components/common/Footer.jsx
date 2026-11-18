import { Link } from 'react-router-dom';
import './Footer.css';

function Footer() {
  return (
    <footer className="footer">
      <div className="footer-container">
        <div className="footer-brand">
          <Link to="/" className="footer-logo">
            <span>✈</span> SkyBook
          </Link>
          <p>
            Making your flight booking experience simple, convenient, and
            enjoyable.
          </p>
        </div>

        <div className="footer-links">
          <h3>Explore</h3>
          <Link to="/">Home</Link>
          <Link to="/about">About Us</Link>
        </div>

        <div className="footer-links">
          <h3>Account</h3>
          <Link to="/login">Login</Link>
          <Link to="/register">Register</Link>
        </div>

        <div className="footer-links">
          <h3>Contact</h3>
          <a href="mailto:support@skybook.com">support@skybook.com</a>
          <p>Vietnam</p>
        </div>
      </div>

      <div className="footer-bottom">
        <p>© {new Date().getFullYear()} SkyBook. All rights reserved.</p>
      </div>
    </footer>
  );
}

export default Footer;
