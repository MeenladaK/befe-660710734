import React, { useState, useEffect } from 'react';
import BookCard from './BookCard';

const FeaturedBooks = () => {
  const [featuredBooks, setFeaturedBooks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const controller = new AbortController();

    (async () => {
      try {
        setLoading(true);

        // ใช้ .env + fallback และตัด / ท้าย URL กันพัง
        const raw = process.env.REACT_APP_API_URL || 'http://localhost:8080';
        const apiUrl = raw.replace(/\/+$/, '');

        const res = await fetch(`${apiUrl}/api/v1/books`, { signal: controller.signal });
        if (!res.ok) throw new Error(`Failed to fetch books (${res.status})`);

        const data = await res.json();

        // สุ่ม 3 เล่ม
        const selected = [...data].sort(() => 0.5 - Math.random()).slice(0, 3);
        setFeaturedBooks(selected);
        setError(null);
      } catch (e) {
        if (e.name !== 'AbortError') setError(e.message);
      } finally {
        setLoading(false);
      }
    })();

    return () => controller.abort();
  }, []);

  if (loading) {
    return (
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div className="text-center py-8 col-span-full">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div className="text-center py-8 col-span-full text-red-600">
          Error: {error}
        </div>
      </div>
    );
  }

  if (featuredBooks.length === 0) {
    return (
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div className="text-center py-8 col-span-full text-gray-600">
          ยังไม่มีหนังสือแนะนำ
        </div>
      </div>
    );
  }

  return (
    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
      {featuredBooks.map((book) => (
        <BookCard key={book.id} book={book} />
      ))}
    </div>
  );
};

export default FeaturedBooks;