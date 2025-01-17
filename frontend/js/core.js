const API_BASE_URL = '/api';
const CACHE = {
    storage: new Map(),
    
    register(key, val, ttl=10) {
        const expirationTime = ttl ? Date.now() + ttl : null;
        this.storage.set(key, { value: val, expirationTime });
    },

    isCached(key){
        const cachedEntry = this.storage.get(key);
        if (cachedEntry && (!cachedEntry.expirationTime || cachedEntry.expirationTime > Date.now())) {
            return cachedEntry;
        }
        return false;
    },

    doCached(key, func, ttl=10) {
        const cachedEntry = this.isCached(key)
        if (cachedEntry) {
            return cachedEntry.value;
        }

        const newValue = func();
        this.register(key, newValue); 
        return newValue;
    },

    invalidate(key) {
        this.storage.delete(key)
    }
};

async function asyncEval(expression) {
    try {
        const func = new Function(`return (async () => { return ${expression}; })()`);
        return await func();
    } catch (error) {
        console.error(`Error evaluating expression: ${expression}`, error);
        throw error;
    }
}

// replaces "Random 1: ${Math.random()}, Random 2: ${Math.random()}, Sum: ${1 + 2}"
// to       "Random 1: 0.8292359135492149, Random 2: 0.9469694489928151, Sum: 3"
async function replaceExpressionsAsync(inputString) {
    const matches = inputString.matchAll(/\${([^}]+)}/g);
    let outputString = inputString;
    let offset = 0;
  
    for (const match of matches) {
        const matchedPart = match[0];
        const originalIndex = match.index;
        const expr = match[1];
        const newIndex = originalIndex + offset;
        let result = await asyncEval(expr);
    
        
        if (result !== undefined && result !== null) {
            if (typeof result == "object")
                result = JSON.stringify(result);
            else
                result = String(result);
        } else {
            result = '';
        }
  
        outputString = outputString.slice(0, newIndex) + result + outputString.slice(newIndex + matchedPart.length);
        offset += result.length - matchedPart.length;
    }

    return outputString;
  }

async function renderTags() {
    for (const element of Array.from(document.getElementsByTagName('render'))) {
        replaceExpressionsAsync(element.innerHTML).then(evaluated => {
            element.innerHTML = evaluated;
        })
    }
}

// Token management
function getToken() {
    return localStorage.getItem('token');
}

function setToken(token) {
    localStorage.setItem('token', token);
}

function removeToken() {
    localStorage.removeItem('token');
}

function isAuthenticated() {
    return !!getToken();
}

// API calls helper
async function apiCall(endpoint, method = 'GET', body = null) {
    const headers = {
        'Content-Type': 'application/json'
    };

    if (isAuthenticated()) {
        headers['Authorization'] = `Bearer ${getToken()}`;
    }

    const options = {
        method,
        headers
    };

    if (body) {
        options.body = JSON.stringify(body);
    }

    const response = await fetch(`${API_BASE_URL}${endpoint}`, options);

    if (response.status === 401) {
        window.location.href = '/login';
        return;
    }

    const data = await response.json();

    if (!response.ok) {
        throw new Error(JSON.stringify(data) || 'Something went wrong');
    }

    return data;
}

function redirectTo(url) {
    window.location.href = url;
}

// Authentication API calls
/**
 * @example
 * login('doctor@example.com', 'password123')
 */
async function login(email, password) {
    const data = await apiCall('/auth/login', 'POST', { email, password });
    setToken(data.token);
    return data;
}

/**
 * @example
 * register({
 *   email: 'patient@example.com',
 *   password: 'password123',
 *   firstName: 'John',
 *   lastName: 'Doe',
 *   role: 'patient'
 * })
 */
async function register(userData) {
    const data = await apiCall('/auth/register', 'POST', userData);
    setToken(data.token);
    return data;
}

async function logout() {
    removeToken();
    window.location.href = '/login';
}

// User profile API calls
/**
 * Returns user profile data:
 * @example Response:
 * {
 *   id: '123',
 *   email: 'user@example.com',
 *   firstName: 'John',
 *   lastName: 'Doe',
 *   role: 'patient'
 * }
 */
async function getUserProfile() {
    return CACHE.doCached("userProfile", async () => {
        return await apiCall('/users/profile')
    })
}

/**
 * @example
 * updateUserProfile({
 *   firstName: 'John',
 *   lastName: 'Smith',
 *   phoneNumber: '1234567890'
 * })
 */
async function updateUserProfile(profileData) {
    CACHE.invalidate("userProfile")
    return await apiCall('/users/profile', 'PUT', profileData);
}

// Appointment API calls
/**
 * @example
 * createAppointment({
 *   doctorId: '123',
 *   date: '2024-01-20',
 *   time: '14:30',
 *   reason: 'Regular checkup',
 *   notes: 'Patient has history of high blood pressure'
 * })
 */
async function createAppointment(appointmentData) {
    return await apiCall('/appointments', 'POST', appointmentData);
}

/**
 * Returns array of appointments:
 * @example Response:
 * [{
 *   id: '456',
 *   doctorId: '123',
 *   patientId: '789',
 *   date: '2024-01-20',
 *   time: '14:30',
 *   status: 'scheduled',
 *   reason: 'Regular checkup',
 *   notes: 'Patient has history of high blood pressure'
 * }]
 */
async function getAppointments() {
    return await apiCall('/appointments?userId=' + (await getUserProfile()).id);
}

/**
 * @example
 * updateAppointment('456', {
 *   date: '2024-01-21',
 *   time: '15:30',
 *   status: 'rescheduled',
 *   notes: 'Appointment rescheduled at patient request'
 * })
 */
async function updateAppointment(id, appointmentData) {
    return await apiCall(`/appointments/${id}`, 'PUT', appointmentData);
}

async function deleteAppointment(id) {
    return await apiCall(`/appointments/${id}`, 'DELETE');
}

// Doctor specific API calls
/**
 * @example
 * updateDoctorExpertise(['expertise_id_1', 'expertise_id_2'])
 */
async function updateDoctorExpertise(expertiseIds) {
    return await apiCall('/doctors/expertises', 'PUT', { expertiseIds });
}

// Admin specific API calls
/**
 * @example
 * createExpertise({
 *   name: 'Cardiology',
 *   description: 'Deals with disorders of the heart'
 * })
 */
async function createExpertise(expertiseData) {
    return await apiCall('/admin/expertises', 'POST', expertiseData);
}

async function getAllExpertises() {
    return CACHE.doCached("expertisesList", async () => {
        return await apiCall('/doctors/expertises')
    }, 30)
}

async function createDoctor(doctorData) {
    return await apiCall('/admin/doctors', 'POST', doctorData);
}

/**
 * @example
 * getAllUsers('doctor') // to get all doctors
 * getAllUsers('patient') // to get all patients
 * getAllUsers() // to get all users
 * 
 * Returns array of users:
 * @example Response:
 * [{
 *   id: '123',
 *   email: 'doctor@example.com',
 *   firstName: 'Jane',
 *   lastName: 'Smith',
 *   role: 'doctor',
 *   expertises: ['Cardiology', 'Internal Medicine']
 * }]
 */
async function getAllUsers() {
    return CACHE.doCached("usersList", async () => {
        return await apiCall('/users')
    }, 30)
}

async function checkIfUserIs(role) {
    return (await getUserProfile()).role === role;
}

async function getUserById(userId) {
    return (await getAllUsers()).find(user => user.id === userId);
}
