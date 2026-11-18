import { NavLink } from 'react-router-dom';

function Navbar() {
  return (
    <header className="navbar">
      <NavLink to="/" className="navbar-logo">
        <span className="logo-icon">✈</span>
        <span>SkyBook</span>
      </NavLink>

      <nav className="navbar-links">
        <NavLink to="/" end className="nav-link">
          Home
        </NavLink>

        <NavLink to="/about" className="nav-link">
          About
        </NavLink>

        <NavLink to="/login" className="nav-link">
          Login
        </NavLink>

        <NavLink to="/register" className="nav-register">
          Register
        </NavLink>
      </nav>
    </header>
  );
}

export default Navbar;
