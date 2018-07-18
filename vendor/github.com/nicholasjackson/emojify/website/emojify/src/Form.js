/*jshint esversion: 6 */ 

import React, { Component } from 'react';

class Form extends Component {
  constructor(props) {
    super(props);
    this.state = {
      urlInput: ""
    };
  }

  onSubmit(e) {
    e.preventDefault();
    console.log('Submit: ' + this.state.urlInput);
  }

  onChange(name, value) {
    this.setState({
      [name]: value.target.value,
    });
  }

  render() {
    return (
      <form onSubmit={this.onSubmit.bind(this)}>
      <div className="row">
        <div className="twelve column">
          <label htmlFor="urlInput">Image URL</label>
          <input className="u-full-width" type="text" placeholder="http://image.com" id="urlInput" onChange={this.onChange.bind(this, 'urlInput')}/>
        </div>
      </div>
      <div className="row">
        <div className="twelve column">
          <input className="button-primary" type="submit" value="Submit"/>
        </div>
      </div>
      </form>
    );
  }
}

export default Form;
