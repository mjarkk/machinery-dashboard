import React from 'react'
import './App.css'
import { get } from './logic/calles'
import Graph from './graph'
// import { DatePicker } from 'antd'

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      data: [],
    }
    this.setup()
  }
  async setup() {
    const res = await get('')
    this.setState({ data: res.data })
  }
  render() {
    return (
      <div className="main">
        {this.state.data.map((el, id) =>
          <Graph key={id} data={el} />
        )}
      </div>
    )
  }
}

export default App;
