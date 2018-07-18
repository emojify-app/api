/*jshint esversion: 6 */ 

import React, { Component } from 'react';
import './App.css';
import Form from './Form.js';

class App extends Component {
  render() {
    return (
      <div className="container">
        <div className="row">
          <div className="twelve column">
            <img className="u-max-full-width" alt="emojify" src="/images/emojify.png"/>
          </div>
        </div>
        <Form/>
      </div>
    );
  }
}

export default App;
