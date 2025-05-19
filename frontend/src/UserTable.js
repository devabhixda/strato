import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { SortFromBottomToTop, SortFromTopToBottom, ListVertical } from '@solar-icons/react';

// Define column configurations for headers and sorting
const columnsConfig = [
  { key: 'humanUser', name: 'User', type: 'string', searchable: true },
  { key: 'createDate', name: 'Create Date', type: 'date', searchable: false },
  { key: 'passwordChangedDate', name: 'Password Changed Date', type: 'date', searchable: false },
  { key: 'daysSinceLastPasswordChange', name: 'Days Since Pwd Change', type: 'number', searchable: false },
  { key: 'lastAccessDate', name: 'Last Access Date', type: 'date', searchable: false },
  { key: 'daysSinceLastAccess', name: 'Days Since Last Access', type: 'number', searchable: false },
  { key: 'mfaEnabled', name: 'MFA Enabled', type: 'string', searchable: false },
];

const dateColumns = columnsConfig.filter(col => col.type === 'date');

// Helper to format date from YYYY-MM-DD input to "Mon DD YYYY" for API
const formatDateForAPI = (isoDateString) => {
  if (!isoDateString) return "";
  const [year, month, day] = isoDateString.split('-');
  const date = new Date(Date.UTC(Number(year), Number(month) - 1, Number(day)));
  
  const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
  return `${monthNames[date.getUTCMonth()]} ${date.getUTCDate()} ${date.getUTCFullYear()}`;
};

// Helper to format Date object to YYYY-MM-DD for date input
const formatDateForInput = (date) => {
    if (!date) return '';
    const year = date.getFullYear();
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const day = date.getDate().toString().padStart(2, '0');
    return `${year}-${month}-${day}`;
};

function UserTable() {
  const [users, setUsers] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [sortConfig, setSortConfig] = useState({ key: 'createDate', direction: 'descending' }); // Default sort

  // Filter states
  const [searchQuery, setSearchQuery] = useState('');
  const [dateFilterColumn, setDateFilterColumn] = useState(dateColumns.length > 0 ? dateColumns[0].key : '');
  const [dateRange, setDateRange] = useState({ startDate: '', endDate: '' });
  const [mfaFilter, setMfaFilter] = useState('all');

  // Add User Modal State
  const [isAddUserModalOpen, setIsAddUserModalOpen] = useState(false);
  const initialNewUserState = {
    humanUser: '',
    createDate: formatDateForInput(new Date()),
    passwordChangedDate: formatDateForInput(new Date()),
    lastAccessDate: formatDateForInput(new Date()),
    mfaEnabled: 'No',
  };
  const [newUser, setNewUser] = useState(initialNewUserState);
  const [addUserError, setAddUserError] = useState('');

  const loadUsers = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch('http://localhost:8080/api/users');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data = await response.json();
      setUsers(data);
      setError(null);
    } catch (e) {
      setError(e.message);
      setUsers([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  const handleNewUserChange = (e) => {
    const { name, value } = e.target;
    setNewUser(prev => ({ ...prev, [name]: value }));
  };

  const handleAddUserSubmit = async (e) => {
    e.preventDefault();
    setAddUserError('');
    if (!newUser.humanUser.trim()) {
        setAddUserError("Username is required.");
        return;
    }

    const userPayload = {
      humanUser: newUser.humanUser,
      createDate: formatDateForAPI(newUser.createDate),
      passwordChangedDate: formatDateForAPI(newUser.passwordChangedDate),
      lastAccessDate: formatDateForAPI(newUser.lastAccessDate),
      mfaEnabled: newUser.mfaEnabled,
    };

    try {
      const response = await fetch('http://localhost:8080/api/users', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(userPayload),
      });
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ message: 'Failed to add user. Server returned an error.' }));
        throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
      }
      setIsAddUserModalOpen(false);
      setNewUser(initialNewUserState);
      loadUsers();
    } catch (err) {
      setAddUserError(err.message || 'An unexpected error occurred.');
    }
  };

  const filteredAndSortedUsers = useMemo(() => {
    let processedUsers = [...users];

    // 1. Apply Search Filter
    if (searchQuery) {
      processedUsers = processedUsers.filter(user =>
        columnsConfig.some(column => {
          if (column.searchable) {
            const value = String(user[column.key]).toLowerCase();
            return value.includes(searchQuery.toLowerCase());
          }
          return false;
        })
      );
    }

    // 2. Apply Date Range Filter
    const { startDate, endDate } = dateRange;
    if (dateFilterColumn && (startDate || endDate)) {
      const start = startDate ? new Date(startDate) : null;
      const end = endDate ? new Date(endDate) : null;

      if (start) start.setHours(0, 0, 0, 0); // Start of the day
      if (end) end.setHours(23, 59, 59, 999); // End of the day

      processedUsers = processedUsers.filter(user => {
        const userDateValue = user[dateFilterColumn];
        if (!userDateValue) return false; // Skip if date is missing
        try {
            const userDate = new Date(userDateValue);
             // Check if userDate is valid
            if (isNaN(userDate.getTime())) {
                console.warn(`Invalid date value for user ${user.humanUser}, column ${dateFilterColumn}: ${userDateValue}`);
                return false;
            }
            if (start && userDate < start) return false;
            if (end && userDate > end) return false;
            return true;
        } catch (e) {
            console.error("Error parsing date for filter:", userDateValue, e);
            return false;
        }
      });
    }

    // 3. Apply MFA Filter
    if (mfaFilter !== 'all') {
      processedUsers = processedUsers.filter(user => user.mfaEnabled === mfaFilter);
    }

    // 4. Apply Sorting
    if (sortConfig.key) {
      const columnToSort = columnsConfig.find(col => col.key === sortConfig.key);
      processedUsers.sort((a, b) => {
        const valA = a[sortConfig.key];
        const valB = b[sortConfig.key];
        let comparison = 0;

        if (columnToSort) {
          switch (columnToSort.type) {
            case 'number':
              comparison = Number(valA) - Number(valB);
              break;
            case 'date':
              try {
                const dateA = new Date(valA);
                const dateB = new Date(valB);
                if (isNaN(dateA.getTime()) && isNaN(dateB.getTime())) comparison = 0;
                else if (isNaN(dateA.getTime())) comparison = 1; // Treat invalid dates as "greater" or last
                else if (isNaN(dateB.getTime())) comparison = -1; // Treat invalid dates as "greater" or last
                else comparison = dateA - dateB;
              } catch (e) { comparison = 0; }
              break;
            case 'string':
            default:
              comparison = String(valA).localeCompare(String(valB));
              break;
          }
        }
        return sortConfig.direction === 'ascending' ? comparison : comparison * -1;
      });
    }
    return processedUsers;
  }, [users, searchQuery, dateFilterColumn, dateRange, mfaFilter, sortConfig]);

  const requestSort = (key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    } else if (sortConfig.key === key && sortConfig.direction === 'descending') {
        direction = 'ascending';
    }
    setSortConfig({ key, direction });
  };

  const getSortIndicator = (columnKey) => {
    const iconSize = 16;
    const iconColor = "currentColor";

    if (sortConfig.key === columnKey) {
      return sortConfig.direction === 'ascending' ?
        <SortFromBottomToTop size={iconSize} color={iconColor} iconStyle="Bold" /> :
        <SortFromTopToBottom size={iconSize} color={iconColor} iconStyle="Bold" />;
    }
    return <ListVertical size={iconSize} color={iconColor} iconStyle="Bold" />;
  };

  const handleDateInputChange = (e) => {
    setDateRange({ ...dateRange, [e.target.name]: e.target.value });
  };
  
  const clearDateFilters = () => {
    setDateRange({ startDate: '', endDate: '' });
    setDateFilterColumn(dateColumns.length > 0 ? dateColumns[0].key : '');
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-screen">
        <div className="animate-spin rounded-full h-16 w-16 border-t-4 border-b-4 border-blue-500"></div>
        <p className="ml-4 text-lg text-gray-700">Loading users...</p>
      </div>
    );
  }

  if (error) {
    return <p className="text-center text-red-500 text-lg p-4">Error loading users: {error}</p>;
  }

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-semibold text-gray-800">User Information Dashboard</h2>
        <button
          onClick={() => { setIsAddUserModalOpen(true); setAddUserError(''); setNewUser(initialNewUserState); }}
          className="bg-indigo-600 hover:bg-indigo-700 text-white font-semibold py-2 px-4 rounded-md shadow-sm transition duration-150 ease-in-out"
        >
          Add User
        </button>
      </div>

      {/* Filters Section */}
      <div className="mb-6 p-4 bg-gray-50 rounded-lg shadow">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 items-end">
          {/* Search Bar */}
          <div>
            <label htmlFor="searchQuery" className="block text-sm font-medium text-gray-700 mb-1">Search Users</label>
            <input
              type="text"
              id="searchQuery"
              name="searchQuery"
              className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
              placeholder="Search by User"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          {/* MFA Filter */}
          <div>
            <label htmlFor="mfaFilter" className="block text-sm font-medium text-gray-700 mb-1">MFA Enabled</label>
            <select
              id="mfaFilter"
              name="mfaFilter"
              className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
              value={mfaFilter}
              onChange={(e) => setMfaFilter(e.target.value)}
            >
              <option value="all">All</option>
              <option value="Yes">Yes</option>
              <option value="No">No</option>
            </select>
          </div>
          
          {/* Date Filter Column Selector */}
          {dateColumns.length > 0 && (
            <div>
              <label htmlFor="dateFilterColumn" className="block text-sm font-medium text-gray-700 mb-1">Filter by Date Column</label>
              <select
                id="dateFilterColumn"
                name="dateFilterColumn"
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                value={dateFilterColumn}
                onChange={(e) => setDateFilterColumn(e.target.value)}
              >
                {dateColumns.map(col => <option key={col.key} value={col.key}>{col.name}</option>)}
              </select>
            </div>
          )}

          {/* Date Range Start */}
          {dateFilterColumn && (
            <div>
              <label htmlFor="startDate" className="block text-sm font-medium text-gray-700 mb-1">Start Date ({columnsConfig.find(c=>c.key === dateFilterColumn)?.name})</label>
              <input
                type="date"
                id="startDate"
                name="startDate"
                className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                value={dateRange.startDate}
                onChange={handleDateInputChange}
              />
            </div>
          )}

          {/* Date Range End */}
          {dateFilterColumn && (
            <div>
              <label htmlFor="endDate" className="block text-sm font-medium text-gray-700 mb-1">End Date ({columnsConfig.find(c=>c.key === dateFilterColumn)?.name})</label>
              <input
                type="date"
                id="endDate"
                name="endDate"
                className="mt-1 block w-full px-3 py-2 bg-white border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                value={dateRange.endDate}
                onChange={handleDateInputChange}
              />
            </div>
          )}
           {dateFilterColumn && (dateRange.startDate || dateRange.endDate) && (
            <div className="pt-5"> {/* Aligns button with inputs */}
                 <button
                    onClick={clearDateFilters}
                    className="w-full bg-red-500 hover:bg-red-600 text-white font-semibold py-2 px-4 rounded-md shadow-sm text-sm"
                >
                    Clear Dates
                </button>
            </div>
           )}
        </div>
      </div>

      <div className="overflow-x-auto shadow-md sm:rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-100">
            <tr>
              {columnsConfig.map((column) => (
                <th
                  key={column.key}
                  scope="col"
                  className="px-6 py-3 text-center text-xs font-medium text-gray-600 uppercase tracking-wider cursor-pointer hover:bg-gray-200 transition-colors duration-150"
                  onClick={() => requestSort(column.key)}
                >
                  <div className="inline-flex items-center">
                    {column.name}
                    <span
                      className="ml-1 text-indigo-600 inline-flex items-center justify-center"
                      style={{ width: '20px', height: '20px' }}
                    >
                      {getSortIndicator(column.key)}
                    </span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {filteredAndSortedUsers.length > 0 ? (
              filteredAndSortedUsers.map((user, index) => (
                // Using a more robust key if available, e.g., user.id. For now, humanUser + index
                <tr key={(user.humanUser || `user-${index}`) + index} className="hover:bg-gray-50 transition-colors duration-150">
                  {columnsConfig.map(column => (
                    <td key={column.key} className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                      {user[column.key]}
                    </td>
                  ))}
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={columnsConfig.length} className="px-6 py-12 text-center text-gray-500 text-lg">
                  No users found matching your criteria.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {users.length > 0 && filteredAndSortedUsers.length === 0 && !loading && (
        <p className="text-center text-gray-600 mt-6">
          No results for the current filters. Try adjusting your search or filter settings.
        </p>
      )}
       <p className="text-sm text-gray-500 mt-4 text-center">
        Displaying {filteredAndSortedUsers.length} of {users.length} users.
      </p>

      {isAddUserModalOpen && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50 flex justify-center items-center p-4">
          <div className="relative bg-white p-8 rounded-lg shadow-xl w-full max-w-md">
            <h3 className="text-xl font-semibold mb-6 text-gray-800">Add New User</h3>
            {addUserError && <p className="mb-4 text-sm text-red-600 bg-red-100 p-2 rounded">{addUserError}</p>}
            <form onSubmit={handleAddUserSubmit} className="space-y-4">
              <div>
                <label htmlFor="humanUser" className="block text-sm font-medium text-gray-700">Username</label>
                <input type="text" name="humanUser" id="humanUser" value={newUser.humanUser} onChange={handleNewUserChange} required
                       className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"/>
              </div>
              <div>
                <label htmlFor="createDate" className="block text-sm font-medium text-gray-700">Create Date</label>
                <input type="date" name="createDate" id="createDate" value={newUser.createDate} onChange={handleNewUserChange} required
                       className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"/>
              </div>
              <div>
                <label htmlFor="passwordChangedDate" className="block text-sm font-medium text-gray-700">Password Changed Date</label>
                <input type="date" name="passwordChangedDate" id="passwordChangedDate" value={newUser.passwordChangedDate} onChange={handleNewUserChange} required
                       className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"/>
              </div>
              <div>
                <label htmlFor="lastAccessDate" className="block text-sm font-medium text-gray-700">Last Access Date</label>
                <input type="date" name="lastAccessDate" id="lastAccessDate" value={newUser.lastAccessDate} onChange={handleNewUserChange} required
                       className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"/>
              </div>
              <div>
                <label htmlFor="mfaEnabled" className="block text-sm font-medium text-gray-700">MFA Enabled</label>
                <select name="mfaEnabled" id="mfaEnabled" value={newUser.mfaEnabled} onChange={handleNewUserChange}
                        className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md">
                  <option value="No">No</option>
                  <option value="Yes">Yes</option>
                </select>
              </div>
              <div className="flex items-center justify-end space-x-3 pt-4">
                <button type="button" onClick={() => setIsAddUserModalOpen(false)}
                        className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md border border-gray-300 transition duration-150 ease-in-out">
                  Cancel
                </button>
                <button type="submit"
                        className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition duration-150 ease-in-out">
                  Add User
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

export default UserTable;