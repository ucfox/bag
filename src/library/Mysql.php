<?php
Class Mysql {
    private $_conn = NULL;

    public function __construct($host, $user, $pwd, $name) {
        $this->_conn = new mysqli($host, $user, $pwd, $name);
        if (mysqli_connect_errno()) {
            throw new Exception("mysql connect failed: " .  mysqli_connect_error());
        }
        if (!$this->_conn->set_charset("utf8")) {
            throw new Exception("mysql character set utf8:" . $this->_conn->error);
        }
    }
    private function _query($sql) {
        return mysqli_query($this->_conn, $sql);
    }

    public function getAll($sql) {
        $query = $this->_query($sql);
        $rs = array();
        if ($query) {
            while($row = mysqli_fetch_assoc($query)) {
                array_push($rs, $row);
            }
            mysqli_free_result($query);
        }
        return $rs;
    }

    public function getOne($sql){
        $query = $this->_query($sql);
        if ($result = @mysqli_fetch_array($query)) {
            return $result[0];
        }
        return NULL;

    }

    public function getRow($sql) {
        $query = $this->_query($sql);
        $rs = array();
        if ($query) {
            $rs = @mysqli_fetch_assoc($query);
        }
        if ($rs) {
            mysqli_free_result($query);
        }
        return $rs;
    }

    public function del($sql) {
        XLogKit::logger('_sql')->info($sql);
        $query = $this->_query($sql);
        if ($query) {
            return @mysqli_fetch_assoc($this->_conn);
        }
        return false;
    }

    public function update($sql) {
        XLogKit::logger('_sql')->info($sql);
        $query = $this->_query($sql);
        if ($query) {
            return @mysqli_affected_rows($this->_conn);
        }
        return false;
    }

    public function insert($sql) {
        XLogKit::logger('_sql')->info($sql);
        $query = $this->_query($sql);
        if ($query) {
            return mysqli_insert_id($this->_conn);
        }
        return false;
    }

    public function begin() {
        $this->_query("BEGIN");
    }

    public function rollback() {
        $this->_query("ROLLBACK");
    }

    public function commit() {
        $this->_query("COMMIT");
    }

    public function __destruct() {
        @mysqli_close($this->_conn);
    }
}
