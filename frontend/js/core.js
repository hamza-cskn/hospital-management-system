const API_BASE_URL = '/api';
const CACHE = {
    storage: new Map(),

    register(key, val, ttl = 10) {
        const expirationTime = ttl ? Date.now() + ttl : null;
        this.storage.set(key, {value: val, expirationTime});
    },

    isCached(key) {
        const cachedEntry = this.storage.get(key);
        if (cachedEntry && (!cachedEntry.expirationTime || cachedEntry.expirationTime > Date.now())) {
            return cachedEntry;
        }
        return false;
    },

    doCached(key, func, ttl = 10) {
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

async function login(email, password) {
    const data = await apiCall('/auth/login', 'POST', {email, password});
    setToken(data.token);
    return data;
}

async function register(userData) {
    const data = await apiCall('/auth/register', 'POST', userData);
    setToken(data.token);
    return data;
}

async function logout() {
    removeToken();
    window.location.href = '/login';
}

async function getUserProfile() {
    return CACHE.doCached("userProfile", async () => {
        return await apiCall('/users/profile')
    })
}

async function updateUserProfile(profileData) {
    CACHE.invalidate("userProfile")
    return await apiCall('/users/profile', 'PUT', profileData);
}

async function createAppointment(appointmentData) {
    return await apiCall('/appointments', 'POST', appointmentData);
}

async function getAppointments() {
    return await apiCall('/appointments?userId=' + (await getUserProfile()).id);
}

async function getAllAppointments() {
    return await apiCall('/appointments');
}

async function updateAppointment(id, appointmentData) {
    return await apiCall(`/appointments/${id}`, 'PUT', appointmentData);
}

async function cancelAppointment(appointmentId) {
    return await updateAppointment(appointmentId, {status: 'cancelled'});
}

async function doneAppointment(appointmentId) {
    return await updateAppointment(appointmentId, {status: 'completed'});
}

async function deleteAppointment(id) {
    return await apiCall(`/appointments/${id}`, 'DELETE');
}

async function updateExpertise(id, expertiseData) {
    return await apiCall('/admin/expertises?id='+id, 'PATCH', expertiseData);
}

async function deleteExpertise(id) {
    return await apiCall('/admin/expertises?id='+id, 'DELETE');
}

async function createExpertise(expertiseData) {
    return await apiCall('/admin/expertises', 'POST', expertiseData);
}

async function getAllExpertises() {
    return CACHE.doCached("expertisesList", async () => {
        return await apiCall('/expertises')
    }, 30)
}

async function createDoctor(doctorData) {
    return await apiCall('/admin/doctors', 'POST', doctorData);
}

async function updateUser(id, doctorData) {
    return await apiCall('/users?userId=' + id, 'PATCH', doctorData);
}

async function deleteUser(id) {
    return await apiCall('/admin/users/' + id, 'DELETE');
}

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

function workStartEnd2WorkPlan(workstart, workend) {
    if (workstart >= workend) {
        alert('Start time must be earlier than end time.');
        return undefined;
    }

    const duration = 24 * 60 * 60 * 1e9;
    const start = new Date(`1970-01-01T${workstart}Z`).getTime() * 1e6;
    const end = new Date(`1970-01-01T${workend}Z`).getTime() * 1e6;
    return {periods: [{startMargin: start, endMargin: end, duration}]}
}

function workPlan2workStartEnd(workPlan) {
    const start = new Date(workPlan.periods[0].startMargin / 1e6).toISOString().slice(11, 16);
    const end = new Date(workPlan.periods[0].endMargin / 1e6).toISOString().slice(11, 16);
    return [start, end];
}

function getDoctorFormDatas() {
    const email = document.getElementById('email').value;
    const firstName = document.getElementById('firstName').value;
    const lastName = document.getElementById('lastName').value;
    const expertises = Array.from(document.getElementById('expertises').selectedOptions).map(option => option.value);
    const workPlan = workStartEnd2WorkPlan(document.getElementById('workstart').value,
        document.getElementById('workend').value)
    return {email, firstName, lastName, expertises, workPlan};
}