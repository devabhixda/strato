import React, { useState, useEffect } from 'react';

function UserTable() {
  const [users, setUsers] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/users');
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setUsers(data);
      } catch (e) {
        setError(e.message);
      } finally {
        setLoading(false);
      }
    };

    fetchUsers();
  }, []); // Empty dependency array means this effect runs once on mount

  if (loading) {
    return <p>Loading users...</p>;
  }

  if (error) {
    return <p>Error loading users: {error}</p>;
  }

  return (
    <div>
      <h2>User Information</h2>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>User</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>Create Date</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>Password Changed Date</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>Days Since Last Password Change</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>Last Access Date</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>Days Since Last Access</th>
            <th style={{ border: '1px solid #ddd', padding: '8px', textAlign: 'left', backgroundColor: '#f2f2f2' }}>MFA Enabled</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user, index) => (
            <tr key={index}> {/* Using index as a key is okay if list is static and items don't reorder */}
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.humanUser}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.createDate}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.passwordChangedDate}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.daysSinceLastPasswordChange}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.lastAccessDate}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.daysSinceLastAccess}</td>
              <td style={{ border: '1px solid #ddd', padding: '8px' }}>{user.mfaEnabled}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default UserTable;