// Test file for shorthand property assignments
export function createUser(name: string, age: number) {
  // Shorthand property assignments: { name, age } is equivalent to { name: name, age: age }
  return { name, age };
}

export function getUserInfo() {
  const username = 'Alice';
  const userAge = 30;
  const isActive = true;

  // Multiple shorthand properties
  return { username, userAge, isActive };
}

// Mixed shorthand and regular properties
export function mixedProperties(id: number) {
  const name = 'Bob';
  const email = 'bob@example.com';

  return {
    id, // shorthand
    name, // shorthand
    email, // shorthand
    timestamp: Date.now(), // regular property
  };
}
